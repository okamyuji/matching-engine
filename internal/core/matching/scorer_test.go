package matching

import (
	"context"
	"math"
	"testing"
)

func TestNewCompositeScorer(t *testing.T) {
	t.Run("valid components", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:      "euclidean",
				Distance:  &EuclideanDistance{Fields: []string{"age"}},
				Weight:    0.5,
				Transform: InverseTransform(),
			},
			{
				Name:       "cosine",
				Similarity: &CosineSimilarity{VectorField: "text"},
				Weight:     0.5,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if len(scorer.Components) != 2 {
			t.Errorf("Expected 2 components, got %d", len(scorer.Components))
		}
	})

	t.Run("no components", func(t *testing.T) {
		_, err := NewCompositeScorer([]ScoringComponent{})
		if err == nil {
			t.Error("Expected error for empty components")
		}
	})

	t.Run("component with neither similarity nor distance", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Weight: 1.0,
			},
		}

		_, err := NewCompositeScorer(components)
		if err == nil {
			t.Error("Expected error for component without similarity or distance")
		}
	})

	t.Run("component with both similarity and distance", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Similarity: &CosineSimilarity{VectorField: "text"},
				Distance:   &EuclideanDistance{Fields: []string{"age"}},
				Weight:     1.0,
			},
		}

		_, err := NewCompositeScorer(components)
		if err == nil {
			t.Error("Expected error for component with both similarity and distance")
		}
	})

	t.Run("invalid weight", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Distance: &EuclideanDistance{Fields: []string{"age"}},
				Weight:   1.5,
			},
		}

		_, err := NewCompositeScorer(components)
		if err == nil {
			t.Error("Expected error for weight > 1")
		}
	})

	t.Run("auto-generated name", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Distance: &EuclideanDistance{Fields: []string{"age"}},
				Weight:   1.0,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if scorer.Components[0].Name != "euclidean" {
			t.Errorf("Expected auto-generated name 'euclidean', got '%s'", scorer.Components[0].Name)
		}
	})
}

func TestCompositeScorer_Score(t *testing.T) {
	t.Run("single component with distance", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:     "age_distance",
				Distance: &EuclideanDistance{Fields: []string{"age"}},
				Weight:   1.0,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetNumerical("age", 0.5)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("age", 0.5)

		score, breakdown, err := scorer.Score(context.Background(), a, b)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Distance is 0, so similarity should be 1 / (1 + 0) = 1.0
		if math.Abs(score-1.0) > 1e-9 {
			t.Errorf("Expected score 1.0 for identical vectors, got %f", score)
		}

		if _, ok := breakdown["age_distance"]; !ok {
			t.Error("Expected breakdown to contain 'age_distance'")
		}
	})

	t.Run("single component with similarity", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:       "cosine",
				Similarity: &CosineSimilarity{VectorField: "text"},
				Weight:     1.0,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{1.0, 2.0, 3.0})

		b := NewFeatureVector("b", "test")
		b.SetEmbedding("text", []float64{1.0, 2.0, 3.0})

		score, breakdown, err := scorer.Score(context.Background(), a, b)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if math.Abs(score-1.0) > 1e-9 {
			t.Errorf("Expected score 1.0, got %f", score)
		}

		if breakdown["cosine"] != 1.0 {
			t.Errorf("Expected breakdown['cosine'] to be 1.0, got %f", breakdown["cosine"])
		}
	})

	t.Run("multiple components with equal weights", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:     "comp1",
				Distance: &ManhattanDistance{Fields: []string{"x"}},
				Weight:   0.5,
			},
			{
				Name:     "comp2",
				Distance: &ManhattanDistance{Fields: []string{"y"}},
				Weight:   0.5,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.0)
		a.SetNumerical("y", 0.0)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 1.0)
		b.SetNumerical("y", 1.0)

		score, breakdown, err := scorer.Score(context.Background(), a, b)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Both distances are 1.0, so similarity is 1/(1+1) = 0.5
		// Average of 0.5 and 0.5 = 0.5
		expected := 0.5
		if math.Abs(score-expected) > 1e-9 {
			t.Errorf("Expected score %f, got %f", expected, score)
		}

		if len(breakdown) != 2 {
			t.Errorf("Expected 2 breakdown entries, got %d", len(breakdown))
		}
	})

	t.Run("multiple components with different weights", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:     "comp1",
				Distance: &ManhattanDistance{Fields: []string{"x"}},
				Weight:   0.75,
			},
			{
				Name:     "comp2",
				Distance: &ManhattanDistance{Fields: []string{"y"}},
				Weight:   0.25,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.0)
		a.SetNumerical("y", 0.0)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 0.0) // distance = 0, similarity = 1.0
		b.SetNumerical("y", 1.0) // distance = 1, similarity = 0.5

		score, breakdown, err := scorer.Score(context.Background(), a, b)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Weighted average: (1.0 * 0.75 + 0.5 * 0.25) / (0.75 + 0.25) = 0.875
		expected := 0.875
		if math.Abs(score-expected) > 1e-9 {
			t.Errorf("Expected score %f, got %f", expected, score)
		}

		if len(breakdown) != 2 {
			t.Errorf("Expected 2 breakdown entries, got %d", len(breakdown))
		}
	})

	t.Run("with transform function", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:      "comp1",
				Distance:  &ManhattanDistance{Fields: []string{"x"}},
				Weight:    1.0,
				Transform: LinearTransform(2.0, 0.0), // double the score
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.0)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 1.0) // distance = 1, base similarity = 0.5

		score, _, err := scorer.Score(context.Background(), a, b)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Base similarity: 1/(1+1) = 0.5
		// After transform: 0.5 * 2 = 1.0
		expected := 1.0
		if math.Abs(score-expected) > 1e-9 {
			t.Errorf("Expected score %f, got %f", expected, score)
		}
	})

	t.Run("with filter - passing", func(t *testing.T) {
		filter, err := CreateFilter(FilterConfig{
			Field:    "age",
			Operator: "gt",
			Value:    0.3,
		})
		if err != nil {
			t.Fatalf("Failed to create filter: %v", err)
		}

		components := []ScoringComponent{
			{
				Name:     "comp1",
				Distance: &ManhattanDistance{Fields: []string{"age"}},
				Weight:   1.0,
				Filter:   filter,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetNumerical("age", 0.5)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("age", 0.5) // passes filter (0.5 > 0.3)

		score, breakdown, err := scorer.Score(context.Background(), a, b)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if score == 0 {
			t.Error("Expected non-zero score when filter passes")
		}

		if len(breakdown) != 1 {
			t.Errorf("Expected 1 breakdown entry, got %d", len(breakdown))
		}
	})

	t.Run("with filter - failing", func(t *testing.T) {
		filter, err := CreateFilter(FilterConfig{
			Field:    "age",
			Operator: "gt",
			Value:    0.7,
		})
		if err != nil {
			t.Fatalf("Failed to create filter: %v", err)
		}

		components := []ScoringComponent{
			{
				Name:     "comp1",
				Distance: &ManhattanDistance{Fields: []string{"age"}},
				Weight:   1.0,
				Filter:   filter,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetNumerical("age", 0.5)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("age", 0.5) // fails filter (0.5 not > 0.7)

		score, breakdown, err := scorer.Score(context.Background(), a, b)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// All components filtered out, score should be 0
		if score != 0 {
			t.Errorf("Expected score 0 when all filters fail, got %f", score)
		}

		// Breakdown should be empty since component was filtered
		if len(breakdown) != 0 {
			t.Errorf("Expected 0 breakdown entries, got %d", len(breakdown))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:       "comp1",
				Similarity: &CosineSimilarity{VectorField: "text"},
				Weight:     1.0,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{1.0, 2.0})

		b := NewFeatureVector("b", "test")
		b.SetEmbedding("text", []float64{1.0, 2.0})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, _, err = scorer.Score(ctx, a, b)
		if err == nil {
			t.Error("Expected error for canceled context")
		}
	})

	t.Run("invalid vector A", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:     "comp1",
				Distance: &ManhattanDistance{Fields: []string{"age"}},
				Weight:   1.0,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("", "test") // Invalid ID
		b := NewFeatureVector("b", "test")

		_, _, err = scorer.Score(context.Background(), a, b)
		if err == nil {
			t.Error("Expected error for invalid vector A")
		}
	})

	t.Run("component computation error", func(t *testing.T) {
		components := []ScoringComponent{
			{
				Name:     "comp1",
				Distance: &EuclideanDistance{Fields: []string{"nonexistent"}},
				Weight:   1.0,
			},
		}

		scorer, err := NewCompositeScorer(components)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		a := NewFeatureVector("a", "test")
		b := NewFeatureVector("b", "test")

		_, _, err = scorer.Score(context.Background(), a, b)
		if err == nil {
			t.Error("Expected error for missing field")
		}
	})
}
