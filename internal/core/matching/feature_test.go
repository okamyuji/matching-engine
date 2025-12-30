package matching

import (
	"math"
	"testing"
)

func TestNewFeatureVector(t *testing.T) {
	fv := NewFeatureVector("test-id", "test-type")

	if fv.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got '%s'", fv.ID)
	}

	if fv.Type != "test-type" {
		t.Errorf("Expected Type to be 'test-type', got '%s'", fv.Type)
	}

	if fv.Numerical == nil {
		t.Error("Expected Numerical map to be initialized")
	}

	if fv.Categorical == nil {
		t.Error("Expected Categorical map to be initialized")
	}

	if fv.Embeddings == nil {
		t.Error("Expected Embeddings map to be initialized")
	}

	if fv.Sparse == nil {
		t.Error("Expected Sparse map to be initialized")
	}

	if fv.TimeSeries == nil {
		t.Error("Expected TimeSeries map to be initialized")
	}

	if fv.Metadata == nil {
		t.Error("Expected Metadata map to be initialized")
	}
}

func TestFeatureVector_SetNumerical(t *testing.T) {
	fv := NewFeatureVector("test", "test")
	fv.SetNumerical("age", 0.5)

	if val, ok := fv.Numerical["age"]; !ok || val != 0.5 {
		t.Errorf("Expected age to be 0.5, got %v", val)
	}
}

func TestFeatureVector_SetCategorical(t *testing.T) {
	fv := NewFeatureVector("test", "test")
	fv.SetCategorical("prefecture", "tokyo", 1.0)

	if cat, ok := fv.Categorical["prefecture"]; !ok {
		t.Error("Expected prefecture category to exist")
	} else if val, ok := cat["tokyo"]; !ok || val != 1.0 {
		t.Errorf("Expected tokyo to be 1.0, got %v", val)
	}
}

func TestFeatureVector_SetEmbedding(t *testing.T) {
	fv := NewFeatureVector("test", "test")
	original := []float64{0.1, 0.2, 0.3}
	fv.SetEmbedding("text", original)

	if emb, ok := fv.Embeddings["text"]; !ok {
		t.Error("Expected text embedding to exist")
	} else {
		if len(emb) != len(original) {
			t.Errorf("Expected embedding length %d, got %d", len(original), len(emb))
		}

		for i := range original {
			if emb[i] != original[i] {
				t.Errorf("Expected emb[%d] to be %f, got %f", i, original[i], emb[i])
			}
		}

		// Verify it's a copy, not the same slice
		original[0] = 999
		if emb[0] == 999 {
			t.Error("Embedding should be a copy, not the same slice")
		}
	}
}

func TestFeatureVector_SetSparse(t *testing.T) {
	fv := NewFeatureVector("test", "test")
	fv.SetSparse("tags", "sports", 0.8)

	if sparse, ok := fv.Sparse["tags"]; !ok {
		t.Error("Expected tags to exist")
	} else if val, ok := sparse["sports"]; !ok || val != 0.8 {
		t.Errorf("Expected sports to be 0.8, got %v", val)
	}
}

func TestFeatureVector_SetTimeSeries(t *testing.T) {
	fv := NewFeatureVector("test", "test")
	stats := &TimeSeriesStats{
		Mean:       100.0,
		Std:        10.0,
		Min:        80.0,
		Max:        120.0,
		Trend:      1.5,
		Volatility: 0.1,
	}
	fv.SetTimeSeries("revenue", stats)

	if ts, ok := fv.TimeSeries["revenue"]; !ok {
		t.Error("Expected revenue time series to exist")
	} else if ts.Mean != 100.0 {
		t.Errorf("Expected mean to be 100.0, got %f", ts.Mean)
	}
}

func TestFeatureVector_SetMetadata(t *testing.T) {
	fv := NewFeatureVector("test", "test")
	fv.SetMetadata("key", "value")

	if val, ok := fv.Metadata["key"]; !ok || val != "value" {
		t.Errorf("Expected metadata key to be 'value', got %v", val)
	}
}

func TestFeatureVector_Validate(t *testing.T) {
	tests := []struct {
		name    string
		fv      *FeatureVector
		wantErr bool
	}{
		{
			name:    "valid feature vector",
			fv:      NewFeatureVector("id", "type"),
			wantErr: false,
		},
		{
			name:    "empty ID",
			fv:      NewFeatureVector("", "type"),
			wantErr: true,
		},
		{
			name:    "empty type",
			fv:      NewFeatureVector("id", ""),
			wantErr: true,
		},
		{
			name:    "both empty",
			fv:      NewFeatureVector("", ""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fv.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeatureVector_Clone(t *testing.T) {
	original := NewFeatureVector("test", "test")
	original.SetNumerical("age", 0.5)
	original.SetCategorical("prefecture", "tokyo", 1.0)
	original.SetEmbedding("text", []float64{0.1, 0.2, 0.3})
	original.SetSparse("tags", "sports", 0.8)
	original.SetTimeSeries("revenue", &TimeSeriesStats{Mean: 100.0})
	original.SetMetadata("key", "value")

	clone := original.Clone()

	// Verify values are the same
	if clone.ID != original.ID {
		t.Error("Clone ID mismatch")
	}
	if clone.Type != original.Type {
		t.Error("Clone Type mismatch")
	}
	if clone.Numerical["age"] != original.Numerical["age"] {
		t.Error("Clone numerical mismatch")
	}
	if clone.Categorical["prefecture"]["tokyo"] != original.Categorical["prefecture"]["tokyo"] {
		t.Error("Clone categorical mismatch")
	}
	if clone.Sparse["tags"]["sports"] != original.Sparse["tags"]["sports"] {
		t.Error("Clone sparse mismatch")
	}

	// Verify independence - modifying clone shouldn't affect original
	clone.SetNumerical("age", 0.9)
	if original.Numerical["age"] == 0.9 {
		t.Error("Clone is not independent - numerical")
	}

	clone.SetCategorical("prefecture", "osaka", 1.0)
	if _, ok := original.Categorical["prefecture"]["osaka"]; ok {
		t.Error("Clone is not independent - categorical")
	}

	clone.Embeddings["text"][0] = 999
	if original.Embeddings["text"][0] == 999 {
		t.Error("Clone is not independent - embeddings")
	}

	clone.SetSparse("tags", "travel", 0.7)
	if _, ok := original.Sparse["tags"]["travel"]; ok {
		t.Error("Clone is not independent - sparse")
	}

	clone.TimeSeries["revenue"].Mean = 200.0
	if original.TimeSeries["revenue"].Mean == 200.0 {
		t.Error("Clone is not independent - time series")
	}
}

func TestNormalizeValue(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		min      float64
		max      float64
		expected float64
	}{
		{"middle value", 50, 0, 100, 0.5},
		{"min value", 0, 0, 100, 0.0},
		{"max value", 100, 0, 100, 1.0},
		{"below min", -10, 0, 100, 0.0},
		{"above max", 110, 0, 100, 1.0},
		{"same min and max", 50, 50, 50, 0.5},
		{"age example", 30, 18, 80, (30.0 - 18.0) / (80.0 - 18.0)},
	}

	const epsilon = 1e-9

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeValue(tt.value, tt.min, tt.max)
			if math.Abs(result-tt.expected) > epsilon {
				t.Errorf("NormalizeValue(%f, %f, %f) = %f, expected %f",
					tt.value, tt.min, tt.max, result, tt.expected)
			}
		})
	}
}

func TestComputeTimeSeriesStats(t *testing.T) {
	t.Run("empty values", func(t *testing.T) {
		stats := ComputeTimeSeriesStats([]float64{})
		if stats.Mean != 0 || stats.Std != 0 || stats.Min != 0 || stats.Max != 0 {
			t.Error("Expected all zero stats for empty values")
		}
	})

	t.Run("single value", func(t *testing.T) {
		stats := ComputeTimeSeriesStats([]float64{100})
		if stats.Mean != 100 {
			t.Errorf("Expected mean 100, got %f", stats.Mean)
		}
		if stats.Min != 100 {
			t.Errorf("Expected min 100, got %f", stats.Min)
		}
		if stats.Max != 100 {
			t.Errorf("Expected max 100, got %f", stats.Max)
		}
	})

	t.Run("multiple values", func(t *testing.T) {
		values := []float64{100, 110, 120, 130, 140}
		stats := ComputeTimeSeriesStats(values)

		if stats.Mean != 120 {
			t.Errorf("Expected mean 120, got %f", stats.Mean)
		}
		if stats.Min != 100 {
			t.Errorf("Expected min 100, got %f", stats.Min)
		}
		if stats.Max != 140 {
			t.Errorf("Expected max 140, got %f", stats.Max)
		}
		if stats.Trend <= 0 {
			t.Errorf("Expected positive trend, got %f", stats.Trend)
		}
	})

	t.Run("constant values", func(t *testing.T) {
		values := []float64{100, 100, 100, 100}
		stats := ComputeTimeSeriesStats(values)

		if stats.Mean != 100 {
			t.Errorf("Expected mean 100, got %f", stats.Mean)
		}
		if stats.Std != 0 {
			t.Errorf("Expected std 0, got %f", stats.Std)
		}
		if stats.Volatility != 0 {
			t.Errorf("Expected volatility 0, got %f", stats.Volatility)
		}
	})
}
