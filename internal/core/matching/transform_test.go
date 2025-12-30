package matching

import (
	"math"
	"testing"
)

func TestLinearTransform(t *testing.T) {
	tests := []struct {
		name     string
		a        float64
		b        float64
		input    float64
		expected float64
	}{
		{"identity", 1.0, 0.0, 5.0, 5.0},
		{"scale by 2", 2.0, 0.0, 3.0, 6.0},
		{"shift by 1", 1.0, 1.0, 2.0, 3.0},
		{"scale and shift", 2.0, 3.0, 4.0, 11.0},
		{"invert", -1.0, 1.0, 0.5, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transform := LinearTransform(tt.a, tt.b)
			result := transform(tt.input)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("LinearTransform(%f, %f)(%f) = %f, expected %f",
					tt.a, tt.b, tt.input, result, tt.expected)
			}
		})
	}
}

func TestInverseTransform(t *testing.T) {
	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"zero distance", 0.0, 1.0},
		{"distance 1", 1.0, 0.5},
		{"distance 4", 4.0, 0.2},
		{"large distance", 99.0, 0.01},
	}

	transform := InverseTransform()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := transform(tt.input)
			if math.Abs(result-tt.expected) > 1e-9 {
				t.Errorf("InverseTransform()(%f) = %f, expected %f",
					tt.input, result, tt.expected)
			}
		})
	}
}

func TestGaussianTransform(t *testing.T) {
	t.Run("centered at mu", func(t *testing.T) {
		transform := GaussianTransform(0.5, 0.1)
		result := transform(0.5)
		if math.Abs(result-1.0) > 1e-9 {
			t.Errorf("Expected 1.0 at center, got %f", result)
		}
	})

	t.Run("symmetric around mu", func(t *testing.T) {
		mu := 0.5
		sigma := 0.1
		delta := 0.05

		transform := GaussianTransform(mu, sigma)
		left := transform(mu - delta)
		right := transform(mu + delta)

		if math.Abs(left-right) > 1e-9 {
			t.Errorf("Expected symmetric values, got left=%f, right=%f", left, right)
		}
	})

	t.Run("decreases away from mu", func(t *testing.T) {
		transform := GaussianTransform(0.5, 0.1)
		center := transform(0.5)
		away := transform(0.7)

		if away >= center {
			t.Errorf("Expected value to decrease away from mu, center=%f, away=%f", center, away)
		}
	})

	t.Run("wider sigma means slower decay", func(t *testing.T) {
		narrow := GaussianTransform(0.5, 0.05)
		wide := GaussianTransform(0.5, 0.2)

		x := 0.6
		narrowValue := narrow(x)
		wideValue := wide(x)

		if narrowValue >= wideValue {
			t.Errorf("Expected wider sigma to have higher value at distance, narrow=%f, wide=%f",
				narrowValue, wideValue)
		}
	})
}

func TestSigmoidTransform(t *testing.T) {
	t.Run("centered at x0", func(t *testing.T) {
		transform := SigmoidTransform(1.0, 0.5)
		result := transform(0.5)
		if math.Abs(result-0.5) > 1e-9 {
			t.Errorf("Expected 0.5 at center, got %f", result)
		}
	})

	t.Run("approaches 0 and 1", func(t *testing.T) {
		transform := SigmoidTransform(1.0, 0.5)

		low := transform(-10.0)
		if low > 0.1 {
			t.Errorf("Expected value near 0 for low input, got %f", low)
		}

		high := transform(10.0)
		if high < 0.9 {
			t.Errorf("Expected value near 1 for high input, got %f", high)
		}
	})

	t.Run("steepness controlled by k", func(t *testing.T) {
		gentle := SigmoidTransform(1.0, 0.5)
		steep := SigmoidTransform(10.0, 0.5)

		x := 0.6
		gentleValue := gentle(x)
		steepValue := steep(x)

		// Steeper function should be closer to 1 when x > x0
		if steepValue <= gentleValue {
			t.Errorf("Expected steeper sigmoid to have higher value at x > x0, gentle=%f, steep=%f",
				gentleValue, steepValue)
		}
	})

	t.Run("monotonic increasing", func(t *testing.T) {
		transform := SigmoidTransform(1.0, 0.5)

		prev := transform(0.0)
		for x := 0.1; x <= 1.0; x += 0.1 {
			curr := transform(x)
			if curr <= prev {
				t.Errorf("Expected monotonic increase, but value at %f (%f) <= value at previous (%f)",
					x, curr, prev)
			}
			prev = curr
		}
	})
}

func TestStepTransform(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
		input     float64
		expected  float64
	}{
		{"below threshold", 0.5, 0.3, 0.0},
		{"at threshold", 0.5, 0.5, 1.0},
		{"above threshold", 0.5, 0.7, 1.0},
		{"far below", 0.5, -10.0, 0.0},
		{"far above", 0.5, 10.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transform := StepTransform(tt.threshold)
			result := transform(tt.input)
			if result != tt.expected {
				t.Errorf("StepTransform(%f)(%f) = %f, expected %f",
					tt.threshold, tt.input, result, tt.expected)
			}
		})
	}
}

func TestTransformChaining(t *testing.T) {
	// Test chaining transforms
	t.Run("inverse then linear", func(t *testing.T) {
		inverse := InverseTransform()
		linear := LinearTransform(2.0, 0.0)

		input := 3.0
		intermediate := inverse(input)   // 1/(1+3) = 0.25
		expected := linear(intermediate) // 0.25 * 2 = 0.5

		result := linear(inverse(input))
		if math.Abs(result-expected) > 1e-9 {
			t.Errorf("Expected chained result %f, got %f", expected, result)
		}
	})
}
