package matching

import "testing"

func TestCreateFilter_GreaterThan(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "gt",
		Value:    0.5,
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("age", 0.6)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.6 > 0.5")
	}

	b.SetNumerical("age", 0.5)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.5 > 0.5")
	}

	b.SetNumerical("age", 0.4)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.4 > 0.5")
	}
}

func TestCreateFilter_LessThan(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "lt",
		Value:    0.5,
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("age", 0.4)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.4 < 0.5")
	}

	b.SetNumerical("age", 0.5)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.5 < 0.5")
	}

	b.SetNumerical("age", 0.6)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.6 < 0.5")
	}
}

func TestCreateFilter_GreaterThanOrEqual(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "gte",
		Value:    0.5,
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("age", 0.6)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.6 >= 0.5")
	}

	b.SetNumerical("age", 0.5)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.5 >= 0.5")
	}

	b.SetNumerical("age", 0.4)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.4 >= 0.5")
	}
}

func TestCreateFilter_LessThanOrEqual(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "lte",
		Value:    0.5,
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("age", 0.4)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.4 <= 0.5")
	}

	b.SetNumerical("age", 0.5)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.5 <= 0.5")
	}

	b.SetNumerical("age", 0.6)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.6 <= 0.5")
	}
}

func TestCreateFilter_Equal_Numerical(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "eq",
		Value:    0.5,
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("age", 0.5)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.5 == 0.5")
	}

	b.SetNumerical("age", 0.6)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.6 == 0.5")
	}
}

func TestCreateFilter_Equal_Categorical(t *testing.T) {
	config := FilterConfig{
		Field:    "prefecture",
		Operator: "eq",
		Value:    "tokyo",
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")

	b1 := NewFeatureVector("b1", "test")
	b1.SetCategorical("prefecture", "tokyo", 1.0)
	if !filter(a, b1) {
		t.Error("Expected filter to pass for tokyo == tokyo")
	}

	b2 := NewFeatureVector("b2", "test")
	b2.SetCategorical("prefecture", "osaka", 1.0)
	if filter(a, b2) {
		t.Error("Expected filter to fail for osaka == tokyo")
	}
}

func TestCreateFilter_NotEqual_Numerical(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "ne",
		Value:    0.5,
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("age", 0.6)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.6 != 0.5")
	}

	b.SetNumerical("age", 0.5)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.5 != 0.5")
	}
}

func TestCreateFilter_NotEqual_Categorical(t *testing.T) {
	config := FilterConfig{
		Field:    "prefecture",
		Operator: "ne",
		Value:    "tokyo",
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")

	b1 := NewFeatureVector("b1", "test")
	b1.SetCategorical("prefecture", "osaka", 1.0)
	if !filter(a, b1) {
		t.Error("Expected filter to pass for osaka != tokyo")
	}

	b2 := NewFeatureVector("b2", "test")
	b2.SetCategorical("prefecture", "tokyo", 1.0)
	if filter(a, b2) {
		t.Error("Expected filter to fail for tokyo != tokyo")
	}
}

func TestCreateFilter_Range(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "range",
		Value: map[string]any{
			"min": 0.3,
			"max": 0.7,
		},
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("age", 0.5)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.5 in [0.3, 0.7]")
	}

	b.SetNumerical("age", 0.3)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.3 in [0.3, 0.7]")
	}

	b.SetNumerical("age", 0.7)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.7 in [0.3, 0.7]")
	}

	b.SetNumerical("age", 0.2)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.2 in [0.3, 0.7]")
	}

	b.SetNumerical("age", 0.8)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.8 in [0.3, 0.7]")
	}
}

func TestCreateFilter_In_Categorical(t *testing.T) {
	config := FilterConfig{
		Field:    "prefecture",
		Operator: "in",
		Value:    []any{"tokyo", "osaka", "kyoto"},
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")

	b1 := NewFeatureVector("b1", "test")
	b1.SetCategorical("prefecture", "tokyo", 1.0)
	if !filter(a, b1) {
		t.Error("Expected filter to pass for tokyo in [tokyo, osaka, kyoto]")
	}

	b2 := NewFeatureVector("b2", "test")
	b2.SetCategorical("prefecture", "osaka", 1.0)
	if !filter(a, b2) {
		t.Error("Expected filter to pass for osaka in [tokyo, osaka, kyoto]")
	}

	b3 := NewFeatureVector("b3", "test")
	b3.SetCategorical("prefecture", "nagoya", 1.0)
	if filter(a, b3) {
		t.Error("Expected filter to fail for nagoya in [tokyo, osaka, kyoto]")
	}
}

func TestCreateFilter_In_Numerical(t *testing.T) {
	config := FilterConfig{
		Field:    "score",
		Operator: "in",
		Value:    []any{0.1, 0.5, 0.9},
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	b.SetNumerical("score", 0.5)
	if !filter(a, b) {
		t.Error("Expected filter to pass for 0.5 in [0.1, 0.5, 0.9]")
	}

	b.SetNumerical("score", 0.3)
	if filter(a, b) {
		t.Error("Expected filter to fail for 0.3 in [0.1, 0.5, 0.9]")
	}
}

func TestCreateFilter_MissingField(t *testing.T) {
	config := FilterConfig{
		Field:    "age",
		Operator: "gt",
		Value:    0.5,
	}

	filter, err := CreateFilter(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	a := NewFeatureVector("a", "test")
	b := NewFeatureVector("b", "test")

	// Don't set the age field
	if filter(a, b) {
		t.Error("Expected filter to fail when field is missing")
	}
}

func TestCreateFilter_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config FilterConfig
	}{
		{
			name: "empty field",
			config: FilterConfig{
				Field:    "",
				Operator: "gt",
				Value:    0.5,
			},
		},
		{
			name: "empty operator",
			config: FilterConfig{
				Field:    "age",
				Operator: "",
				Value:    0.5,
			},
		},
		{
			name: "unknown operator",
			config: FilterConfig{
				Field:    "age",
				Operator: "unknown",
				Value:    0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := CreateFilter(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}
