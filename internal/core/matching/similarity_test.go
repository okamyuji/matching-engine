package matching

import (
	"math"
	"testing"
)

func TestCosineSimilarity_Compute(t *testing.T) {
	t.Run("identical vectors", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{1.0, 2.0, 3.0})

		b := NewFeatureVector("b", "test")
		b.SetEmbedding("text", []float64{1.0, 2.0, 3.0})

		sim := &CosineSimilarity{VectorField: "text"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if math.Abs(result-1.0) > 1e-9 {
			t.Errorf("Expected similarity 1.0, got %f", result)
		}
	})

	t.Run("orthogonal vectors", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{1.0, 0.0})

		b := NewFeatureVector("b", "test")
		b.SetEmbedding("text", []float64{0.0, 1.0})

		sim := &CosineSimilarity{VectorField: "text"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if math.Abs(result) > 1e-9 {
			t.Errorf("Expected similarity 0.0, got %f", result)
		}
	})

	t.Run("opposite vectors", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{1.0, 2.0, 3.0})

		b := NewFeatureVector("b", "test")
		b.SetEmbedding("text", []float64{-1.0, -2.0, -3.0})

		sim := &CosineSimilarity{VectorField: "text"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if math.Abs(result-(-1.0)) > 1e-9 {
			t.Errorf("Expected similarity -1.0, got %f", result)
		}
	})

	t.Run("zero vector", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{0.0, 0.0, 0.0})

		b := NewFeatureVector("b", "test")
		b.SetEmbedding("text", []float64{1.0, 2.0, 3.0})

		sim := &CosineSimilarity{VectorField: "text"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 0.0 {
			t.Errorf("Expected similarity 0.0 for zero vector, got %f", result)
		}
	})

	t.Run("dimension mismatch", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{1.0, 2.0})

		b := NewFeatureVector("b", "test")
		b.SetEmbedding("text", []float64{1.0, 2.0, 3.0})

		sim := &CosineSimilarity{VectorField: "text"}
		_, err := sim.Compute(a, b)

		if err == nil {
			t.Error("Expected error for dimension mismatch")
		}
	})

	t.Run("missing field", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetEmbedding("text", []float64{1.0, 2.0})

		b := NewFeatureVector("b", "test")

		sim := &CosineSimilarity{VectorField: "text"}
		_, err := sim.Compute(a, b)

		if err == nil {
			t.Error("Expected error for missing field")
		}
	})

	t.Run("empty field name", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		b := NewFeatureVector("b", "test")

		sim := &CosineSimilarity{VectorField: ""}
		_, err := sim.Compute(a, b)

		if err == nil {
			t.Error("Expected error for empty field name")
		}
	})
}

func TestCosineSimilarity_Name(t *testing.T) {
	sim := &CosineSimilarity{}
	if sim.Name() != "cosine" {
		t.Errorf("Expected name 'cosine', got '%s'", sim.Name())
	}
}

func TestJaccardSimilarity_Compute(t *testing.T) {
	t.Run("identical sets", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetSparse("tags", "sports", 1.0)
		a.SetSparse("tags", "travel", 1.0)

		b := NewFeatureVector("b", "test")
		b.SetSparse("tags", "sports", 1.0)
		b.SetSparse("tags", "travel", 1.0)

		sim := &JaccardSimilarity{SparseField: "tags"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if math.Abs(result-1.0) > 1e-9 {
			t.Errorf("Expected similarity 1.0, got %f", result)
		}
	})

	t.Run("partial overlap", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetSparse("tags", "sports", 1.0)
		a.SetSparse("tags", "travel", 1.0)

		b := NewFeatureVector("b", "test")
		b.SetSparse("tags", "sports", 1.0)
		b.SetSparse("tags", "music", 1.0)

		sim := &JaccardSimilarity{SparseField: "tags"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Intersection: {sports} = 1
		// Union: {sports, travel, music} = 3
		// Jaccard = 1/3
		expected := 1.0 / 3.0
		if math.Abs(result-expected) > 1e-9 {
			t.Errorf("Expected similarity %f, got %f", expected, result)
		}
	})

	t.Run("no overlap", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetSparse("tags", "sports", 1.0)

		b := NewFeatureVector("b", "test")
		b.SetSparse("tags", "music", 1.0)

		sim := &JaccardSimilarity{SparseField: "tags"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 0.0 {
			t.Errorf("Expected similarity 0.0, got %f", result)
		}
	})

	t.Run("both empty sets", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.Sparse["tags"] = make(map[string]float64)

		b := NewFeatureVector("b", "test")
		b.Sparse["tags"] = make(map[string]float64)

		sim := &JaccardSimilarity{SparseField: "tags"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 1.0 {
			t.Errorf("Expected similarity 1.0 for both empty sets, got %f", result)
		}
	})

	t.Run("one empty set", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.Sparse["tags"] = make(map[string]float64)

		b := NewFeatureVector("b", "test")
		b.SetSparse("tags", "sports", 1.0)

		sim := &JaccardSimilarity{SparseField: "tags"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 0.0 {
			t.Errorf("Expected similarity 0.0 for one empty set, got %f", result)
		}
	})

	t.Run("missing field", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		b := NewFeatureVector("b", "test")

		sim := &JaccardSimilarity{SparseField: "tags"}
		_, err := sim.Compute(a, b)

		if err == nil {
			t.Error("Expected error for missing field")
		}
	})
}

func TestJaccardSimilarity_Name(t *testing.T) {
	sim := &JaccardSimilarity{}
	if sim.Name() != "jaccard" {
		t.Errorf("Expected name 'jaccard', got '%s'", sim.Name())
	}
}

func TestCategoricalSimilarity_Compute(t *testing.T) {
	t.Run("matching categories", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetCategorical("prefecture", "tokyo", 1.0)

		b := NewFeatureVector("b", "test")
		b.SetCategorical("prefecture", "tokyo", 1.0)

		sim := &CategoricalSimilarity{Field: "prefecture"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 1.0 {
			t.Errorf("Expected similarity 1.0, got %f", result)
		}
	})

	t.Run("different categories", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetCategorical("prefecture", "tokyo", 1.0)

		b := NewFeatureVector("b", "test")
		b.SetCategorical("prefecture", "osaka", 1.0)

		sim := &CategoricalSimilarity{Field: "prefecture"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 0.0 {
			t.Errorf("Expected similarity 0.0, got %f", result)
		}
	})

	t.Run("multiple categories with max", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetCategorical("prefecture", "tokyo", 0.7)
		a.SetCategorical("prefecture", "osaka", 0.3)

		b := NewFeatureVector("b", "test")
		b.SetCategorical("prefecture", "tokyo", 0.9)
		b.SetCategorical("prefecture", "kyoto", 0.1)

		sim := &CategoricalSimilarity{Field: "prefecture"}
		result, err := sim.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 1.0 {
			t.Errorf("Expected similarity 1.0 (both max is tokyo), got %f", result)
		}
	})

	t.Run("missing field", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		b := NewFeatureVector("b", "test")

		sim := &CategoricalSimilarity{Field: "prefecture"}
		_, err := sim.Compute(a, b)

		if err == nil {
			t.Error("Expected error for missing field")
		}
	})

	t.Run("empty field name", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		b := NewFeatureVector("b", "test")

		sim := &CategoricalSimilarity{Field: ""}
		_, err := sim.Compute(a, b)

		if err == nil {
			t.Error("Expected error for empty field name")
		}
	})
}

func TestCategoricalSimilarity_Name(t *testing.T) {
	sim := &CategoricalSimilarity{}
	if sim.Name() != "categorical" {
		t.Errorf("Expected name 'categorical', got '%s'", sim.Name())
	}
}
