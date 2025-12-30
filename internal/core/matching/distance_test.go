package matching

import (
	"math"
	"testing"
)

func TestEuclideanDistance_Compute(t *testing.T) {
	t.Run("identical vectors", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.5)
		a.SetNumerical("y", 0.3)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 0.5)
		b.SetNumerical("y", 0.3)

		dist := &EuclideanDistance{Fields: []string{"x", "y"}}
		result, err := dist.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 0 {
			t.Errorf("Expected distance 0, got %f", result)
		}
	})

	t.Run("different vectors", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.0)
		a.SetNumerical("y", 0.0)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 3.0)
		b.SetNumerical("y", 4.0)

		dist := &EuclideanDistance{Fields: []string{"x", "y"}}
		result, err := dist.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := 5.0 // sqrt(3^2 + 4^2) = 5
		if math.Abs(result-expected) > 1e-9 {
			t.Errorf("Expected distance %f, got %f", expected, result)
		}
	})

	t.Run("missing field in vector A", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.5)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 0.5)
		b.SetNumerical("y", 0.3)

		dist := &EuclideanDistance{Fields: []string{"x", "y"}}
		_, err := dist.Compute(a, b)

		if err == nil {
			t.Error("Expected error for missing field in vector A")
		}
	})

	t.Run("missing field in vector B", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.5)
		a.SetNumerical("y", 0.3)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 0.5)

		dist := &EuclideanDistance{Fields: []string{"x", "y"}}
		_, err := dist.Compute(a, b)

		if err == nil {
			t.Error("Expected error for missing field in vector B")
		}
	})

	t.Run("empty fields", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		b := NewFeatureVector("b", "test")

		dist := &EuclideanDistance{Fields: []string{}}
		_, err := dist.Compute(a, b)

		if err == nil {
			t.Error("Expected error for empty fields")
		}
	})

	t.Run("single field", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("age", 0.3)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("age", 0.5)

		dist := &EuclideanDistance{Fields: []string{"age"}}
		result, err := dist.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := 0.2
		if math.Abs(result-expected) > 1e-9 {
			t.Errorf("Expected distance %f, got %f", expected, result)
		}
	})
}

func TestEuclideanDistance_Name(t *testing.T) {
	dist := &EuclideanDistance{}
	if dist.Name() != "euclidean" {
		t.Errorf("Expected name 'euclidean', got '%s'", dist.Name())
	}
}

func TestManhattanDistance_Compute(t *testing.T) {
	t.Run("identical vectors", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.5)
		a.SetNumerical("y", 0.3)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 0.5)
		b.SetNumerical("y", 0.3)

		dist := &ManhattanDistance{Fields: []string{"x", "y"}}
		result, err := dist.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if result != 0 {
			t.Errorf("Expected distance 0, got %f", result)
		}
	})

	t.Run("different vectors", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 1.0)
		a.SetNumerical("y", 2.0)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 4.0)
		b.SetNumerical("y", 6.0)

		dist := &ManhattanDistance{Fields: []string{"x", "y"}}
		result, err := dist.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := 7.0 // |1-4| + |2-6| = 3 + 4 = 7
		if math.Abs(result-expected) > 1e-9 {
			t.Errorf("Expected distance %f, got %f", expected, result)
		}
	})

	t.Run("negative differences", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 5.0)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 2.0)

		dist := &ManhattanDistance{Fields: []string{"x"}}
		result, err := dist.Compute(a, b)

		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		expected := 3.0
		if math.Abs(result-expected) > 1e-9 {
			t.Errorf("Expected distance %f, got %f", expected, result)
		}
	})

	t.Run("missing field", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		a.SetNumerical("x", 0.5)

		b := NewFeatureVector("b", "test")
		b.SetNumerical("x", 0.5)

		dist := &ManhattanDistance{Fields: []string{"x", "y"}}
		_, err := dist.Compute(a, b)

		if err == nil {
			t.Error("Expected error for missing field")
		}
	})

	t.Run("empty fields", func(t *testing.T) {
		a := NewFeatureVector("a", "test")
		b := NewFeatureVector("b", "test")

		dist := &ManhattanDistance{Fields: []string{}}
		_, err := dist.Compute(a, b)

		if err == nil {
			t.Error("Expected error for empty fields")
		}
	})
}

func TestManhattanDistance_Name(t *testing.T) {
	dist := &ManhattanDistance{}
	if dist.Name() != "manhattan" {
		t.Errorf("Expected name 'manhattan', got '%s'", dist.Name())
	}
}
