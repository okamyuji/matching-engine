package matching

import (
	"testing"
)

func TestNewRanker(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		config := &RankingConfig{}
		ranker := NewRanker(config)

		if ranker.Config.SortOrder != "desc" {
			t.Errorf("Expected default sort order 'desc', got '%s'", ranker.Config.SortOrder)
		}

		if ranker.Config.Limit != 20 {
			t.Errorf("Expected default limit 20, got %d", ranker.Config.Limit)
		}
	})

	t.Run("custom values", func(t *testing.T) {
		config := &RankingConfig{
			SortOrder: "asc",
			Limit:     10,
			Offset:    5,
		}
		ranker := NewRanker(config)

		if ranker.Config.SortOrder != "asc" {
			t.Errorf("Expected sort order 'asc', got '%s'", ranker.Config.SortOrder)
		}

		if ranker.Config.Limit != 10 {
			t.Errorf("Expected limit 10, got %d", ranker.Config.Limit)
		}
	})

	t.Run("random factor clamping", func(t *testing.T) {
		config := &RankingConfig{
			RandomFactor: 1.5,
		}
		ranker := NewRanker(config)

		if ranker.Config.RandomFactor != 1.0 {
			t.Errorf("Expected random factor clamped to 1.0, got %f", ranker.Config.RandomFactor)
		}

		config.RandomFactor = -0.5
		ranker = NewRanker(config)

		if ranker.Config.RandomFactor != 0.0 {
			t.Errorf("Expected random factor clamped to 0.0, got %f", ranker.Config.RandomFactor)
		}
	})
}

func TestRanker_Rank_Basic(t *testing.T) {
	t.Run("empty matches", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{})
		result := ranker.Rank([]ScoredMatch{})

		if len(result) != 0 {
			t.Errorf("Expected empty result for empty input")
		}
	})

	t.Run("sort descending", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder: "desc",
			Limit:     10,
		})

		matches := []ScoredMatch{
			{Candidate: NewFeatureVector("a", "test"), Score: 0.5},
			{Candidate: NewFeatureVector("b", "test"), Score: 0.8},
			{Candidate: NewFeatureVector("c", "test"), Score: 0.3},
			{Candidate: NewFeatureVector("d", "test"), Score: 0.9},
		}

		result := ranker.Rank(matches)

		if len(result) != 4 {
			t.Errorf("Expected 4 results, got %d", len(result))
		}

		// Check descending order
		if result[0].Score != 0.9 {
			t.Errorf("Expected first score 0.9, got %f", result[0].Score)
		}
		if result[1].Score != 0.8 {
			t.Errorf("Expected second score 0.8, got %f", result[1].Score)
		}
		if result[2].Score != 0.5 {
			t.Errorf("Expected third score 0.5, got %f", result[2].Score)
		}
		if result[3].Score != 0.3 {
			t.Errorf("Expected fourth score 0.3, got %f", result[3].Score)
		}

		// Check ranks
		for i, match := range result {
			if match.Rank != i+1 {
				t.Errorf("Expected rank %d, got %d", i+1, match.Rank)
			}
		}
	})

	t.Run("sort ascending", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder: "asc",
			Limit:     10,
		})

		matches := []ScoredMatch{
			{Candidate: NewFeatureVector("a", "test"), Score: 0.5},
			{Candidate: NewFeatureVector("b", "test"), Score: 0.8},
			{Candidate: NewFeatureVector("c", "test"), Score: 0.3},
		}

		result := ranker.Rank(matches)

		// Check ascending order
		if result[0].Score != 0.3 {
			t.Errorf("Expected first score 0.3, got %f", result[0].Score)
		}
		if result[1].Score != 0.5 {
			t.Errorf("Expected second score 0.5, got %f", result[1].Score)
		}
		if result[2].Score != 0.8 {
			t.Errorf("Expected third score 0.8, got %f", result[2].Score)
		}
	})
}

func TestRanker_Rank_Pagination(t *testing.T) {
	t.Run("with limit", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder: "desc",
			Limit:     2,
		})

		matches := []ScoredMatch{
			{Candidate: NewFeatureVector("a", "test"), Score: 0.9},
			{Candidate: NewFeatureVector("b", "test"), Score: 0.8},
			{Candidate: NewFeatureVector("c", "test"), Score: 0.7},
			{Candidate: NewFeatureVector("d", "test"), Score: 0.6},
		}

		result := ranker.Rank(matches)

		if len(result) != 2 {
			t.Errorf("Expected 2 results due to limit, got %d", len(result))
		}

		if result[0].Score != 0.9 || result[1].Score != 0.8 {
			t.Error("Expected top 2 scores")
		}
	})

	t.Run("with offset", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder: "desc",
			Limit:     2,
			Offset:    1,
		})

		matches := []ScoredMatch{
			{Candidate: NewFeatureVector("a", "test"), Score: 0.9},
			{Candidate: NewFeatureVector("b", "test"), Score: 0.8},
			{Candidate: NewFeatureVector("c", "test"), Score: 0.7},
			{Candidate: NewFeatureVector("d", "test"), Score: 0.6},
		}

		result := ranker.Rank(matches)

		if len(result) != 2 {
			t.Errorf("Expected 2 results, got %d", len(result))
		}

		if result[0].Score != 0.8 || result[1].Score != 0.7 {
			t.Error("Expected scores 0.8 and 0.7 with offset 1")
		}

		// Ranks should still be 1, 2 (not 2, 3)
		if result[0].Rank != 1 || result[1].Rank != 2 {
			t.Errorf("Expected ranks 1, 2, got %d, %d", result[0].Rank, result[1].Rank)
		}
	})

	t.Run("offset beyond length", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder: "desc",
			Limit:     10,
			Offset:    100,
		})

		matches := []ScoredMatch{
			{Candidate: NewFeatureVector("a", "test"), Score: 0.9},
		}

		result := ranker.Rank(matches)

		if len(result) != 0 {
			t.Errorf("Expected 0 results when offset > length, got %d", len(result))
		}
	})
}

func TestRanker_Rank_Diversity(t *testing.T) {
	t.Run("diversity enabled", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder: "desc",
			Diversity: &DiversityConfig{
				Enabled:      true,
				GroupByField: "category",
				MaxPerGroup:  2,
			},
			Limit: 10,
		})

		matches := []ScoredMatch{
			{
				Candidate: createVectorWithCategory("a", "tech"),
				Score:     0.9,
			},
			{
				Candidate: createVectorWithCategory("b", "tech"),
				Score:     0.85,
			},
			{
				Candidate: createVectorWithCategory("c", "tech"),
				Score:     0.8,
			},
			{
				Candidate: createVectorWithCategory("d", "finance"),
				Score:     0.75,
			},
			{
				Candidate: createVectorWithCategory("e", "finance"),
				Score:     0.7,
			},
		}

		result := ranker.Rank(matches)

		// Should have max 2 from each category
		techCount := 0
		financeCount := 0

		for _, match := range result {
			cat := match.Candidate.Categorical["category"]
			if cat["tech"] > 0 {
				techCount++
			} else if cat["finance"] > 0 {
				financeCount++
			}
		}

		if techCount > 2 {
			t.Errorf("Expected max 2 tech items, got %d", techCount)
		}

		if financeCount > 2 {
			t.Errorf("Expected max 2 finance items, got %d", financeCount)
		}

		// Should have 4 total (2 tech + 2 finance)
		if len(result) != 4 {
			t.Errorf("Expected 4 results with diversity, got %d", len(result))
		}
	})

	t.Run("diversity disabled", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder: "desc",
			Diversity: &DiversityConfig{
				Enabled:      false,
				GroupByField: "category",
				MaxPerGroup:  1,
			},
			Limit: 10,
		})

		matches := []ScoredMatch{
			{Candidate: createVectorWithCategory("a", "tech"), Score: 0.9},
			{Candidate: createVectorWithCategory("b", "tech"), Score: 0.8},
			{Candidate: createVectorWithCategory("c", "tech"), Score: 0.7},
		}

		result := ranker.Rank(matches)

		// Should return all 3 since diversity is disabled
		if len(result) != 3 {
			t.Errorf("Expected 3 results when diversity disabled, got %d", len(result))
		}
	})
}

func TestRanker_Rank_Randomness(t *testing.T) {
	t.Run("randomness applied", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder:    "desc",
			RandomFactor: 0.5,
			Limit:        10,
		})

		matches := []ScoredMatch{
			{Candidate: NewFeatureVector("a", "test"), Score: 0.9},
			{Candidate: NewFeatureVector("b", "test"), Score: 0.8},
			{Candidate: NewFeatureVector("c", "test"), Score: 0.7},
			{Candidate: NewFeatureVector("d", "test"), Score: 0.6},
		}

		// Run multiple times to check if order varies (due to randomness)
		// Note: This test is probabilistic
		firstResult := ranker.Rank(matches)

		// Just verify it returns correct number of results
		// and ranks are assigned
		if len(firstResult) != 4 {
			t.Errorf("Expected 4 results, got %d", len(firstResult))
		}

		for i, match := range firstResult {
			if match.Rank != i+1 {
				t.Errorf("Expected rank %d, got %d", i+1, match.Rank)
			}
		}
	})

	t.Run("no randomness", func(t *testing.T) {
		ranker := NewRanker(&RankingConfig{
			SortOrder:    "desc",
			RandomFactor: 0.0,
			Limit:        10,
		})

		matches := []ScoredMatch{
			{Candidate: NewFeatureVector("a", "test"), Score: 0.9},
			{Candidate: NewFeatureVector("b", "test"), Score: 0.8},
		}

		result1 := ranker.Rank(matches)
		result2 := ranker.Rank(matches)

		// Without randomness, results should be identical
		if result1[0].Candidate.ID != result2[0].Candidate.ID {
			t.Error("Expected identical ordering without randomness")
		}
	})
}

// Helper function to create a vector with a category
func createVectorWithCategory(id, category string) *FeatureVector {
	fv := NewFeatureVector(id, "test")
	fv.SetCategorical("category", category, 1.0)
	return fv
}
