package matching

import (
	"context"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestNewConfigurableEngine(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		config := &MatchingConfig{
			Version:     "1.0",
			Domain:      "test",
			Description: "Test config",
			Scoring: ScoringConfig{
				MinScore: 0.3,
				Components: []ComponentConfig{
					{
						Name:   "age_distance",
						Type:   "euclidean",
						Fields: []string{"age"},
						Weight: 1.0,
					},
				},
			},
			Ranking: RankingConfig{
				SortOrder: "desc",
				Limit:     10,
			},
		}

		engine, err := NewConfigurableEngine(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if engine.config != config {
			t.Error("Expected engine config to match input config")
		}

		if engine.scorer == nil {
			t.Error("Expected scorer to be initialized")
		}

		if engine.ranker == nil {
			t.Error("Expected ranker to be initialized")
		}
	})

	t.Run("invalid configuration", func(t *testing.T) {
		config := &MatchingConfig{
			Version: "", // Missing version
			Domain:  "test",
		}

		_, err := NewConfigurableEngine(config)
		if err == nil {
			t.Error("Expected error for invalid configuration")
		}
	})
}

func TestNewConfigurableEngineFromFile(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.json")

	configJSON := `{
  "version": "1.0",
  "domain": "test",
  "description": "Test config",
  "scoring": {
    "min_score": 0.3,
    "components": [
      {
        "name": "age_similarity",
        "type": "euclidean",
        "fields": ["age"],
        "weight": 1.0
      }
    ]
  },
  "ranking": {
    "sort_order": "desc",
    "limit": 10
  }
}`

	if err := os.WriteFile(configPath, []byte(configJSON), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	engine, err := NewConfigurableEngineFromFile(configPath)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if engine == nil {
		t.Fatal("Expected engine to be created")
	}

	if engine.config.Version != "1.0" {
		t.Errorf("Expected version '1.0', got '%s'", engine.config.Version)
	}
}

func TestConfigurableEngine_FindMatches(t *testing.T) {
	t.Run("basic matching", func(t *testing.T) {
		config := &MatchingConfig{
			Version:     "1.0",
			Domain:      "test",
			Description: "Test",
			Scoring: ScoringConfig{
				MinScore: 0.0,
				Components: []ComponentConfig{
					{
						Name:   "age_similarity",
						Type:   "euclidean",
						Fields: []string{"age"},
						Weight: 1.0,
					},
				},
			},
			Ranking: RankingConfig{
				SortOrder: "desc",
				Limit:     10,
			},
		}

		engine, err := NewConfigurableEngine(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		source := NewFeatureVector("source", "test")
		source.SetNumerical("age", 0.5)

		candidates := []*FeatureVector{
			createCandidate("c1", 0.5), // Exact match
			createCandidate("c2", 0.6), // Close
			createCandidate("c3", 0.3), // Far
		}

		matches, err := engine.FindMatches(context.Background(), source, candidates)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(matches) != 3 {
			t.Errorf("Expected 3 matches, got %d", len(matches))
		}

		// c1 should be first (exact match, highest score)
		if matches[0].Candidate.ID != "c1" {
			t.Errorf("Expected c1 to be first, got %s", matches[0].Candidate.ID)
		}

		// Check ranks
		for i, match := range matches {
			if match.Rank != i+1 {
				t.Errorf("Expected rank %d, got %d", i+1, match.Rank)
			}
		}
	})

	t.Run("with minimum score filter", func(t *testing.T) {
		config := &MatchingConfig{
			Version:     "1.0",
			Domain:      "test",
			Description: "Test",
			Scoring: ScoringConfig{
				MinScore: 0.5, // Only return matches with score >= 0.5
				Components: []ComponentConfig{
					{
						Name:   "age_similarity",
						Type:   "euclidean",
						Fields: []string{"age"},
						Weight: 1.0,
					},
				},
			},
			Ranking: RankingConfig{
				SortOrder: "desc",
				Limit:     10,
			},
		}

		engine, err := NewConfigurableEngine(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		source := NewFeatureVector("source", "test")
		source.SetNumerical("age", 0.5)

		candidates := []*FeatureVector{
			createCandidate("c1", 0.5),  // score = 1.0 (pass)
			createCandidate("c2", 0.55), // score ~ 0.95 (pass)
			createCandidate("c3", 0.9),  // score ~ 0.71 (pass)
			createCandidate("c4", 2.0),  // score ~ 0.4 (fail)
		}

		matches, err := engine.FindMatches(context.Background(), source, candidates)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Should filter out c4
		if len(matches) > 3 {
			t.Errorf("Expected at most 3 matches after min_score filter, got %d", len(matches))
		}

		// All returned matches should have score >= 0.5
		for _, match := range matches {
			if match.Score < 0.5 {
				t.Errorf("Expected all scores >= 0.5, got %f for %s", match.Score, match.Candidate.ID)
			}
		}
	})

	t.Run("with limit", func(t *testing.T) {
		config := &MatchingConfig{
			Version:     "1.0",
			Domain:      "test",
			Description: "Test",
			Scoring: ScoringConfig{
				MinScore: 0.0,
				Components: []ComponentConfig{
					{
						Name:   "age_similarity",
						Type:   "euclidean",
						Fields: []string{"age"},
						Weight: 1.0,
					},
				},
			},
			Ranking: RankingConfig{
				SortOrder: "desc",
				Limit:     2, // Only return top 2
			},
		}

		engine, err := NewConfigurableEngine(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		source := NewFeatureVector("source", "test")
		source.SetNumerical("age", 0.5)

		candidates := []*FeatureVector{
			createCandidate("c1", 0.5),
			createCandidate("c2", 0.6),
			createCandidate("c3", 0.55),
			createCandidate("c4", 0.7),
		}

		matches, err := engine.FindMatches(context.Background(), source, candidates)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(matches) != 2 {
			t.Errorf("Expected 2 matches due to limit, got %d", len(matches))
		}
	})

	t.Run("multi-component scoring", func(t *testing.T) {
		config := &MatchingConfig{
			Version:     "1.0",
			Domain:      "test",
			Description: "Test",
			Scoring: ScoringConfig{
				MinScore: 0.0,
				Components: []ComponentConfig{
					{
						Name:   "numerical",
						Type:   "euclidean",
						Fields: []string{"x"},
						Weight: 0.5,
					},
					{
						Name:   "categorical",
						Type:   "categorical",
						Field:  "category",
						Weight: 0.5,
					},
				},
			},
			Ranking: RankingConfig{
				SortOrder: "desc",
				Limit:     10,
			},
		}

		engine, err := NewConfigurableEngine(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		source := NewFeatureVector("source", "test")
		source.SetNumerical("x", 0.5)
		source.SetCategorical("category", "A", 1.0)

		c1 := NewFeatureVector("c1", "test")
		c1.SetNumerical("x", 0.5)               // exact match
		c1.SetCategorical("category", "A", 1.0) // exact match
		// Expected score: (1.0 * 0.5 + 1.0 * 0.5) / 1.0 = 1.0

		c2 := NewFeatureVector("c2", "test")
		c2.SetNumerical("x", 0.5)               // exact match
		c2.SetCategorical("category", "B", 1.0) // no match
		// Expected score: (1.0 * 0.5 + 0.0 * 0.5) / 1.0 = 0.5

		matches, err := engine.FindMatches(context.Background(), source, []*FeatureVector{c1, c2})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(matches) != 2 {
			t.Fatalf("Expected 2 matches, got %d", len(matches))
		}

		// c1 should be first (higher score)
		if matches[0].Candidate.ID != "c1" {
			t.Errorf("Expected c1 first, got %s", matches[0].Candidate.ID)
		}

		// Check breakdown
		if len(matches[0].Breakdown) != 2 {
			t.Errorf("Expected breakdown with 2 components, got %d", len(matches[0].Breakdown))
		}
	})

	t.Run("invalid source vector", func(t *testing.T) {
		config := &MatchingConfig{
			Version:     "1.0",
			Domain:      "test",
			Description: "Test",
			Scoring: ScoringConfig{
				MinScore: 0.0,
				Components: []ComponentConfig{
					{
						Name:   "age_similarity",
						Type:   "euclidean",
						Fields: []string{"age"},
						Weight: 1.0,
					},
				},
			},
			Ranking: RankingConfig{
				SortOrder: "desc",
				Limit:     10,
			},
		}

		engine, err := NewConfigurableEngine(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		source := NewFeatureVector("", "test") // Invalid ID

		_, err = engine.FindMatches(context.Background(), source, []*FeatureVector{})
		if err == nil {
			t.Error("Expected error for invalid source vector")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		config := &MatchingConfig{
			Version:     "1.0",
			Domain:      "test",
			Description: "Test",
			Scoring: ScoringConfig{
				MinScore: 0.0,
				Components: []ComponentConfig{
					{
						Name:   "age_similarity",
						Type:   "euclidean",
						Fields: []string{"age"},
						Weight: 1.0,
					},
				},
			},
			Ranking: RankingConfig{
				SortOrder: "desc",
				Limit:     10,
			},
		}

		engine, err := NewConfigurableEngine(config)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		source := NewFeatureVector("source", "test")
		source.SetNumerical("age", 0.5)

		candidates := []*FeatureVector{
			createCandidate("c1", 0.5),
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err = engine.FindMatches(ctx, source, candidates)
		if err == nil {
			t.Error("Expected error for canceled context")
		}
	})
}

func TestConfigurableEngine_Config(t *testing.T) {
	config := &MatchingConfig{
		Version:     "1.0",
		Domain:      "test",
		Description: "Test",
		Scoring: ScoringConfig{
			MinScore: 0.3,
			Components: []ComponentConfig{
				{
					Name:   "age_similarity",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 1.0,
				},
			},
		},
		Ranking: RankingConfig{
			SortOrder: "desc",
			Limit:     10,
		},
	}

	engine, err := NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retrievedConfig := engine.Config()
	if retrievedConfig != config {
		t.Error("Expected retrieved config to match original")
	}
}

// Helper function to create a candidate with a single numerical feature
func createCandidate(id string, age float64) *FeatureVector {
	fv := NewFeatureVector(id, "test")
	fv.SetNumerical("age", age)
	return fv
}

func TestIntegrationExample(t *testing.T) {
	// This is a comprehensive integration test demonstrating full engine usage
	config := &MatchingConfig{
		Version:     "1.0",
		Domain:      "dating",
		Description: "Dating matching engine",
		Scoring: ScoringConfig{
			MinScore: 0.4,
			Components: []ComponentConfig{
				{
					Name:   "age_compatibility",
					Type:   "euclidean",
					Fields: []string{"age"},
					Weight: 0.3,
					Transform: &TransformConfig{
						Type: "gaussian",
						Params: map[string]float64{
							"mu":    0.0,
							"sigma": 0.1,
						},
					},
				},
				{
					Name:   "location_match",
					Type:   "categorical",
					Field:  "prefecture",
					Weight: 0.3,
				},
				{
					Name:   "interests_overlap",
					Type:   "jaccard",
					Field:  "interests",
					Weight: 0.4,
				},
			},
		},
		Ranking: RankingConfig{
			SortOrder: "desc",
			Diversity: &DiversityConfig{
				Enabled:      true,
				GroupByField: "prefecture",
				MaxPerGroup:  2,
			},
			RandomFactor: 0.1,
			Limit:        5,
		},
	}

	engine, err := NewConfigurableEngine(config)
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Create source user
	source := NewFeatureVector("user1", "dating_user")
	source.SetNumerical("age", 0.4) // 28 years old (normalized)
	source.SetCategorical("prefecture", "tokyo", 1.0)
	source.SetSparse("interests", "sports", 1.0)
	source.SetSparse("interests", "travel", 1.0)
	source.SetSparse("interests", "music", 1.0)

	// Create candidate users
	c1 := NewFeatureVector("user2", "dating_user")
	c1.SetNumerical("age", 0.42)
	c1.SetCategorical("prefecture", "tokyo", 1.0)
	c1.SetSparse("interests", "sports", 1.0)
	c1.SetSparse("interests", "travel", 1.0)

	c2 := NewFeatureVector("user3", "dating_user")
	c2.SetNumerical("age", 0.6)
	c2.SetCategorical("prefecture", "osaka", 1.0)
	c2.SetSparse("interests", "music", 1.0)

	c3 := NewFeatureVector("user4", "dating_user")
	c3.SetNumerical("age", 0.38)
	c3.SetCategorical("prefecture", "tokyo", 1.0)
	c3.SetSparse("interests", "sports", 1.0)
	c3.SetSparse("interests", "gaming", 1.0)

	candidates := []*FeatureVector{c1, c2, c3}

	matches, err := engine.FindMatches(context.Background(), source, candidates)
	if err != nil {
		t.Fatalf("Failed to find matches: %v", err)
	}

	// Verify results
	if len(matches) == 0 {
		t.Fatal("Expected at least one match")
	}

	// Check that scores are reasonable
	for i, match := range matches {
		t.Logf("Match %d: %s (score: %.3f)", i+1, match.Candidate.ID, match.Score)

		if match.Score < 0 || match.Score > 1 {
			t.Errorf("Score out of range [0,1]: %f", match.Score)
		}

		if match.Rank != i+1 {
			t.Errorf("Expected rank %d, got %d", i+1, match.Rank)
		}

		if len(match.Breakdown) != 3 {
			t.Errorf("Expected 3 components in breakdown, got %d", len(match.Breakdown))
		}
	}

	// Verify descending score order
	for i := 1; i < len(matches); i++ {
		if matches[i].Score > matches[i-1].Score+1e-6 { // Allow small floating point errors
			// Note: With randomness, this might not always hold
			// so we just log it rather than fail
			t.Logf("Warning: Scores not in strict descending order (randomness may be applied)")
		}
	}
}

func TestBuildTransform(t *testing.T) {
	tests := []struct {
		name        string
		config      *TransformConfig
		input       float64
		expectError bool
		checkOutput func(float64) bool
	}{
		{
			name: "linear transform",
			config: &TransformConfig{
				Type:   "linear",
				Params: map[string]float64{"a": 2.0, "b": 1.0},
			},
			input:       1.0,
			expectError: false,
			checkOutput: func(output float64) bool {
				return math.Abs(output-3.0) < 1e-9 // 2*1 + 1 = 3
			},
		},
		{
			name: "inverse transform",
			config: &TransformConfig{
				Type: "inverse",
			},
			input:       1.0,
			expectError: false,
			checkOutput: func(output float64) bool {
				return math.Abs(output-0.5) < 1e-9 // 1/(1+1) = 0.5
			},
		},
		{
			name: "unknown transform type",
			config: &TransformConfig{
				Type: "unknown",
			},
			input:       1.0,
			expectError: true,
		},
		{
			name: "linear missing params",
			config: &TransformConfig{
				Type:   "linear",
				Params: map[string]float64{"a": 2.0}, // Missing 'b'
			},
			input:       1.0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transform, err := buildTransform(tt.config)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output := transform(tt.input)
			if tt.checkOutput != nil && !tt.checkOutput(output) {
				t.Errorf("Transform output check failed for input %f, got %f", tt.input, output)
			}
		})
	}
}
