package matching

import "fmt"

// FilterFunc フィルタ関数。ペアがフィルタを通過する場合trueを返す
// フィルタを通過しないペアはマッチングから除外される（スコア0）
type FilterFunc func(*FeatureVector, *FeatureVector) bool

// FilterConfig フィルタ条件を定義する
type FilterConfig struct {
	Field    string `json:"field"`    // フィールド名
	Operator string `json:"operator"` // 演算子
	Value    any    `json:"value"`    // 比較値
}

// CreateFilter 設定からフィルタ関数を作成する
func CreateFilter(config FilterConfig) (FilterFunc, error) {
	if config.Field == "" {
		return nil, fmt.Errorf("%w: field is required", ErrInvalidFilter)
	}

	if config.Operator == "" {
		return nil, fmt.Errorf("%w: operator is required", ErrInvalidFilter)
	}

	switch config.Operator {
	case "gt":
		return createGreaterThanFilter(config)
	case "lt":
		return createLessThanFilter(config)
	case "gte":
		return createGreaterThanOrEqualFilter(config)
	case "lte":
		return createLessThanOrEqualFilter(config)
	case "eq":
		return createEqualFilter(config)
	case "ne":
		return createNotEqualFilter(config)
	case "range":
		return createRangeFilter(config)
	case "in":
		return createInFilter(config)
	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidOperator, config.Operator)
	}
}

func createGreaterThanFilter(config FilterConfig) (FilterFunc, error) {
	threshold, ok := config.Value.(float64)
	if !ok {
		return nil, fmt.Errorf("%w: gt operator requires float64 value", ErrInvalidFilter)
	}

	return func(a, b *FeatureVector) bool {
		val, ok := b.Numerical[config.Field]
		if !ok {
			return false
		}
		return val > threshold
	}, nil
}

func createLessThanFilter(config FilterConfig) (FilterFunc, error) {
	threshold, ok := config.Value.(float64)
	if !ok {
		return nil, fmt.Errorf("%w: lt operator requires float64 value", ErrInvalidFilter)
	}

	return func(a, b *FeatureVector) bool {
		val, ok := b.Numerical[config.Field]
		if !ok {
			return false
		}
		return val < threshold
	}, nil
}

func createGreaterThanOrEqualFilter(config FilterConfig) (FilterFunc, error) {
	threshold, ok := config.Value.(float64)
	if !ok {
		return nil, fmt.Errorf("%w: gte operator requires float64 value", ErrInvalidFilter)
	}

	return func(a, b *FeatureVector) bool {
		val, ok := b.Numerical[config.Field]
		if !ok {
			return false
		}
		return val >= threshold
	}, nil
}

func createLessThanOrEqualFilter(config FilterConfig) (FilterFunc, error) {
	threshold, ok := config.Value.(float64)
	if !ok {
		return nil, fmt.Errorf("%w: lte operator requires float64 value", ErrInvalidFilter)
	}

	return func(a, b *FeatureVector) bool {
		val, ok := b.Numerical[config.Field]
		if !ok {
			return false
		}
		return val <= threshold
	}, nil
}

func createEqualFilter(config FilterConfig) (FilterFunc, error) {
	if config.Value == nil {
		return nil, fmt.Errorf("%w: eq operator requires a value", ErrInvalidFilter)
	}

	// Try as float64 first
	if threshold, ok := config.Value.(float64); ok {
		return func(a, b *FeatureVector) bool {
			val, ok := b.Numerical[config.Field]
			if !ok {
				return false
			}
			return val == threshold
		}, nil
	}

	// Try as string (for categorical)
	if str, ok := config.Value.(string); ok {
		return func(a, b *FeatureVector) bool {
			cat, ok := b.Categorical[config.Field]
			if !ok {
				return false
			}
			// Check if the category exists and has value > 0
			if val, ok := cat[str]; ok && val > 0 {
				return true
			}
			return false
		}, nil
	}

	return nil, fmt.Errorf("%w: eq operator requires float64 or string value", ErrInvalidFilter)
}

func createNotEqualFilter(config FilterConfig) (FilterFunc, error) {
	if config.Value == nil {
		return nil, fmt.Errorf("%w: ne operator requires a value", ErrInvalidFilter)
	}

	// Try as float64 first
	if threshold, ok := config.Value.(float64); ok {
		return func(a, b *FeatureVector) bool {
			val, ok := b.Numerical[config.Field]
			if !ok {
				return false
			}
			return val != threshold
		}, nil
	}

	// Try as string (for categorical)
	if str, ok := config.Value.(string); ok {
		return func(a, b *FeatureVector) bool {
			cat, ok := b.Categorical[config.Field]
			if !ok {
				return true // Field doesn't exist, so it's not equal
			}
			// Check if the category doesn't exist or has value 0
			if val, ok := cat[str]; !ok || val == 0 {
				return true
			}
			return false
		}, nil
	}

	return nil, fmt.Errorf("%w: ne operator requires float64 or string value", ErrInvalidFilter)
}

func createRangeFilter(config FilterConfig) (FilterFunc, error) {
	rangeMap, ok := config.Value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: range operator requires map with 'min' and 'max'", ErrInvalidFilter)
	}

	minVal, minOk := rangeMap["min"].(float64)
	maxVal, maxOk := rangeMap["max"].(float64)

	if !minOk || !maxOk {
		return nil, fmt.Errorf("%w: range operator requires float64 'min' and 'max'", ErrInvalidFilter)
	}

	return func(a, b *FeatureVector) bool {
		val, ok := b.Numerical[config.Field]
		if !ok {
			return false
		}
		return val >= minVal && val <= maxVal
	}, nil
}

func createInFilter(config FilterConfig) (FilterFunc, error) {
	// Try as []any first
	if values, ok := config.Value.([]any); ok {
		// Check if they're all strings (for categorical)
		strValues := make([]string, 0, len(values))
		allStrings := true
		for _, v := range values {
			if str, ok := v.(string); ok {
				strValues = append(strValues, str)
			} else {
				allStrings = false
				break
			}
		}

		if allStrings {
			valueSet := make(map[string]bool, len(strValues))
			for _, v := range strValues {
				valueSet[v] = true
			}

			return func(a, b *FeatureVector) bool {
				cat, ok := b.Categorical[config.Field]
				if !ok {
					return false
				}
				// Check if any of the categories in the set is present with value > 0
				for category, val := range cat {
					if valueSet[category] && val > 0 {
						return true
					}
				}
				return false
			}, nil
		}

		// Try as float64 values
		floatValues := make([]float64, 0, len(values))
		for _, v := range values {
			if f, ok := v.(float64); ok {
				floatValues = append(floatValues, f)
			} else {
				return nil, fmt.Errorf("%w: in operator requires array of float64 or string", ErrInvalidFilter)
			}
		}

		valueSet := make(map[float64]bool, len(floatValues))
		for _, v := range floatValues {
			valueSet[v] = true
		}

		return func(a, b *FeatureVector) bool {
			val, ok := b.Numerical[config.Field]
			if !ok {
				return false
			}
			return valueSet[val]
		}, nil
	}

	return nil, fmt.Errorf("%w: in operator requires array value", ErrInvalidFilter)
}
