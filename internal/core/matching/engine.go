package matching

import (
	"context"
	"fmt"
)

// ConfigurableEngine JSONファイルから設定される主要なマッチングエンジン
type ConfigurableEngine struct {
	config *MatchingConfig
	scorer *CompositeScorer
	ranker *Ranker
}

// NewConfigurableEngine 設定から新しいエンジンを作成する
func NewConfigurableEngine(config *MatchingConfig) (*ConfigurableEngine, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, err
	}

	scorer, err := BuildScorer(&config.Scoring)
	if err != nil {
		return nil, fmt.Errorf("failed to build scorer: %w", err)
	}

	ranker, err := BuildRanker(&config.Ranking)
	if err != nil {
		return nil, fmt.Errorf("failed to build ranker: %w", err)
	}

	return &ConfigurableEngine{
		config: config,
		scorer: scorer,
		ranker: ranker,
	}, nil
}

// NewConfigurableEngineFromFile 設定ファイルから新しいエンジンを作成する
func NewConfigurableEngineFromFile(path string) (*ConfigurableEngine, error) {
	config, err := LoadConfig(path)
	if err != nil {
		return nil, err
	}

	return NewConfigurableEngine(config)
}

// FindMatches ソース特徴ベクトルに対して候補とのマッチを検索してランク付けする
func (e *ConfigurableEngine) FindMatches(
	ctx context.Context,
	source *FeatureVector,
	candidates []*FeatureVector,
) ([]ScoredMatch, error) {
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("source vector validation failed: %w", err)
	}

	matches := make([]ScoredMatch, 0, len(candidates))

	for _, candidate := range candidates {
		// コンテキストキャンセルチェック
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if err := candidate.Validate(); err != nil {
			// 無効な候補はスキップ
			continue
		}

		// スコア計算
		score, breakdown, err := e.scorer.Score(ctx, source, candidate)
		if err != nil {
			// スコアリング失敗した候補はスキップ
			continue
		}

		// 最小スコアフィルタ適用
		if score < e.config.Scoring.MinScore {
			continue
		}

		matches = append(matches, ScoredMatch{
			Candidate: candidate,
			Score:     score,
			Breakdown: breakdown,
		})
	}

	// マッチをランク付け
	rankedMatches := e.ranker.Rank(matches)

	return rankedMatches, nil
}

// Config エンジンの設定を返す
func (e *ConfigurableEngine) Config() *MatchingConfig {
	return e.config
}
