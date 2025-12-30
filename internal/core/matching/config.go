package matching

import (
	"encoding/json"
	"fmt"
	"os"
)

// MatchingConfig 完全なマッチング設定を表す
type MatchingConfig struct {
	Version     string        `json:"version"`
	Domain      string        `json:"domain"`
	Description string        `json:"description"`
	Scoring     ScoringConfig `json:"scoring"`
	Ranking     RankingConfig `json:"ranking"`
}

// ScoringConfig スコアリング設定を定義する
type ScoringConfig struct {
	MinScore   float64           `json:"min_score"`
	Components []ComponentConfig `json:"components"`
}

// ComponentConfig 単一のスコアリングコンポーネントを定義する
type ComponentConfig struct {
	Name      string           `json:"name"`
	Type      string           `json:"type"`   // "euclidean", "manhattan", "cosine", "jaccard", "categorical"
	Fields    []string         `json:"fields"` // 複数フィールド関数用
	Field     string           `json:"field"`  // 単一フィールド関数用
	Weight    float64          `json:"weight"`
	Transform *TransformConfig `json:"transform"`
	Filter    *FilterConfig    `json:"filter"`
}

// TransformConfig 変換設定を定義する
type TransformConfig struct {
	Type   string             `json:"type"`   // "linear", "inverse", "gaussian", "sigmoid", "step"
	Params map[string]float64 `json:"params"` // 変換のパラメータ
}

// LoadConfig JSONファイルから設定を読み込む
func LoadConfig(path string) (*MatchingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config MatchingConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// ValidateConfig 設定を検証する
func ValidateConfig(config *MatchingConfig) error {
	if config.Version == "" {
		return fmt.Errorf("%w: version is required", ErrInvalidConfig)
	}

	if config.Domain == "" {
		return fmt.Errorf("%w: domain is required", ErrInvalidConfig)
	}

	if len(config.Scoring.Components) == 0 {
		return ErrNoComponents
	}

	// 各コンポーネントを検証
	for i, comp := range config.Scoring.Components {
		if comp.Type == "" {
			return fmt.Errorf("%w: component %d missing type", ErrInvalidConfig, i)
		}

		if comp.Weight < 0 || comp.Weight > 1 {
			return fmt.Errorf("%w: component %d has weight %f", ErrInvalidWeight, i, comp.Weight)
		}

		// タイプに基づいてフィールドを検証
		switch comp.Type {
		case "euclidean", "manhattan":
			if len(comp.Fields) == 0 {
				return fmt.Errorf("%w: component %d of type %s requires fields", ErrInvalidConfig, i, comp.Type)
			}
		case "cosine", "jaccard", "categorical":
			if comp.Field == "" {
				return fmt.Errorf("%w: component %d of type %s requires field", ErrInvalidConfig, i, comp.Type)
			}
		default:
			return fmt.Errorf("%w: unknown component type %s", ErrInvalidConfig, comp.Type)
		}
	}

	return nil
}

// BuildScorer 設定からCompositeScorerを作成する
func BuildScorer(config *ScoringConfig) (*CompositeScorer, error) {
	components := make([]ScoringComponent, 0, len(config.Components))

	for _, compConfig := range config.Components {
		comp := ScoringComponent{
			Name:   compConfig.Name,
			Weight: compConfig.Weight,
		}

		// タイプに基づいて類似度または距離関数を作成
		switch compConfig.Type {
		case "euclidean":
			comp.Distance = &EuclideanDistance{Fields: compConfig.Fields}
		case "manhattan":
			comp.Distance = &ManhattanDistance{Fields: compConfig.Fields}
		case "cosine":
			comp.Similarity = &CosineSimilarity{VectorField: compConfig.Field}
		case "jaccard":
			comp.Similarity = &JaccardSimilarity{SparseField: compConfig.Field}
		case "categorical":
			comp.Similarity = &CategoricalSimilarity{Field: compConfig.Field}
		default:
			return nil, fmt.Errorf("%w: unknown component type %s", ErrInvalidConfig, compConfig.Type)
		}

		// 変換が指定されている場合は追加
		if compConfig.Transform != nil {
			transform, err := buildTransform(compConfig.Transform)
			if err != nil {
				return nil, err
			}
			comp.Transform = transform
		}

		// フィルタが指定されている場合は追加
		if compConfig.Filter != nil {
			filter, err := CreateFilter(*compConfig.Filter)
			if err != nil {
				return nil, err
			}
			comp.Filter = filter
		}

		components = append(components, comp)
	}

	return NewCompositeScorer(components)
}

// buildTransform 設定からTransformFuncを作成する
func buildTransform(config *TransformConfig) (TransformFunc, error) {
	switch config.Type {
	case "linear":
		a, aOk := config.Params["a"]
		b, bOk := config.Params["b"]
		if !aOk || !bOk {
			return nil, fmt.Errorf("%w: linear transform requires 'a' and 'b' params", ErrInvalidTransform)
		}
		return LinearTransform(a, b), nil

	case "inverse":
		return InverseTransform(), nil

	case "gaussian":
		mu, muOk := config.Params["mu"]
		sigma, sigmaOk := config.Params["sigma"]
		if !muOk || !sigmaOk {
			return nil, fmt.Errorf("%w: gaussian transform requires 'mu' and 'sigma' params", ErrInvalidTransform)
		}
		return GaussianTransform(mu, sigma), nil

	case "sigmoid":
		k, kOk := config.Params["k"]
		x0, x0Ok := config.Params["x0"]
		if !kOk || !x0Ok {
			return nil, fmt.Errorf("%w: sigmoid transform requires 'k' and 'x0' params", ErrInvalidTransform)
		}
		return SigmoidTransform(k, x0), nil

	case "step":
		threshold, ok := config.Params["threshold"]
		if !ok {
			return nil, fmt.Errorf("%w: step transform requires 'threshold' param", ErrInvalidTransform)
		}
		return StepTransform(threshold), nil

	default:
		return nil, fmt.Errorf("%w: %s", ErrInvalidTransform, config.Type)
	}
}

// BuildRanker 設定からRankerを作成する
func BuildRanker(config *RankingConfig) (*Ranker, error) {
	return NewRanker(config), nil
}
