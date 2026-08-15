package matching

import (
	"context"
	"fmt"
)

// CompositeScorer 複数のスコアリング関数を組み合わせて最終スコアを計算する
type CompositeScorer struct {
	Components []ScoringComponent
}

// ScoringComponent 単一のスコアリングコンポーネントを表す
type ScoringComponent struct {
	// Name このコンポーネントの名前（内訳レポート用）
	Name string

	// Similarity 類似度関数または Distance 距離関数（いずれか一方のみ設定）
	Similarity SimilarityFunction
	Distance   DistanceFunction

	// Weight このコンポーネントの重み (0-1)
	Weight float64

	// Transform オプションの変換関数
	Transform TransformFunc

	// Filter オプションのフィルタ関数
	Filter FilterFunc
}

// NewCompositeScorer 指定されたコンポーネントで新しい複合スコアラーを作成する
func NewCompositeScorer(components []ScoringComponent) (*CompositeScorer, error) {
	if len(components) == 0 {
		return nil, ErrNoComponents
	}

	// コンポーネント検証
	for i, comp := range components {
		if comp.Similarity == nil && comp.Distance == nil {
			return nil, fmt.Errorf("%w: component %d has neither similarity nor distance function", ErrInvalidConfig, i)
		}

		if comp.Similarity != nil && comp.Distance != nil {
			return nil, fmt.Errorf("%w: component %d has both similarity and distance function", ErrInvalidConfig, i)
		}

		if comp.Weight < 0 || comp.Weight > 1 {
			return nil, fmt.Errorf("%w: component %d has weight %f", ErrInvalidWeight, i, comp.Weight)
		}

		// 名前が指定されていない場合はデフォルト名を設定
		if comp.Name == "" {
			if comp.Similarity != nil {
				components[i].Name = comp.Similarity.Name()
			} else {
				components[i].Name = comp.Distance.Name()
			}
		}
	}

	return &CompositeScorer{
		Components: components,
	}, nil
}

// Score 2つの特徴ベクトル間の複合スコアを計算する
// 戻り値:
// - finalScore: 全コンポーネントスコアの加重平均
// - breakdown: コンポーネント名から個別スコアへのマップ
// - error: いずれかのコンポーネントが失敗した場合
func (s *CompositeScorer) Score(
	ctx context.Context,
	a, b *FeatureVector,
) (float64, map[string]float64, error) {
	if err := a.Validate(); err != nil {
		return 0, nil, fmt.Errorf("vector A validation failed: %w", err)
	}

	if err := b.Validate(); err != nil {
		return 0, nil, fmt.Errorf("vector B validation failed: %w", err)
	}

	var totalScore float64
	var totalWeight float64
	breakdown := make(map[string]float64)

	for _, comp := range s.Components {
		// コンテキストキャンセルチェック
		select {
		case <-ctx.Done():
			return 0, nil, ctx.Err()
		default:
		}

		// 1. フィルタチェック
		if comp.Filter != nil && !comp.Filter(a, b) {
			// フィルタを通過しない場合はこのコンポーネントをスキップ
			continue
		}

		// 2. 類似度または距離を計算
		score, err := comp.rawScore(a, b)
		if err != nil {
			return 0, nil, err
		}

		// 3. 変換関数を適用
		if comp.Transform != nil {
			score = comp.Transform(score)
		}

		// 4. 重みを適用
		weightedScore := score * comp.Weight
		totalScore += weightedScore
		totalWeight += comp.Weight

		// 内訳に保存（明確性のため重み付け前の値）
		breakdown[comp.Name] = score
	}

	// 5. 総重みで正規化
	if totalWeight == 0 {
		// 全コンポーネントがフィルタで除外された
		return 0, breakdown, nil
	}

	finalScore := totalScore / totalWeight

	// [0, 1]にクランプ
	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > 1 {
		finalScore = 1
	}

	return finalScore, breakdown, nil
}

// rawScore 類似度関数または距離関数から変換前のスコア（0〜1）を計算する。
// 距離は 1 / (1 + d) で類似度に変換する
func (c *ScoringComponent) rawScore(a, b *FeatureVector) (float64, error) {
	if c.Similarity != nil {
		score, err := c.Similarity.Compute(a, b)
		if err != nil {
			return 0, fmt.Errorf("component %s similarity computation failed: %w", c.Name, err)
		}
		return score, nil
	}
	if c.Distance != nil {
		distance, err := c.Distance.Compute(a, b)
		if err != nil {
			return 0, fmt.Errorf("component %s distance computation failed: %w", c.Name, err)
		}
		return 1.0 / (1.0 + distance), nil
	}
	return 0, nil
}
