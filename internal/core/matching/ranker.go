package matching

import (
	"math/rand"
	"sort"
)

// ScoredMatch スコア付きマッチを表す
type ScoredMatch struct {
	Candidate *FeatureVector     `json:"candidate"` // 候補特徴ベクトル
	Score     float64            `json:"score"`     // 総合スコア
	Breakdown map[string]float64 `json:"breakdown"` // コンポーネント別スコア内訳
	Rank      int                `json:"rank"`      // 結果セット内の順位（1始まり）
}

// Ranker マッチング結果のランキング戦略を処理する
type Ranker struct {
	Config *RankingConfig
}

// RankingConfig ランキング設定を定義する
type RankingConfig struct {
	SortOrder    string           `json:"sort_order"`    // "desc" または "asc"
	Diversity    *DiversityConfig `json:"diversity"`     // 多様性設定
	RandomFactor float64          `json:"random_factor"` // ランダム係数 (0-1)
	Limit        int              `json:"limit"`         // 返す結果の数
	Offset       int              `json:"offset"`        // ページネーション用オフセット
}

// DiversityConfig 多様性設定を定義する
type DiversityConfig struct {
	Enabled      bool   `json:"enabled"`        // 多様性の有効/無効
	GroupByField string `json:"group_by_field"` // グループ化するフィールド
	MaxPerGroup  int    `json:"max_per_group"`  // グループあたりの最大アイテム数
}

// NewRanker 指定された設定で新しいランカーを作成する
func NewRanker(config *RankingConfig) *Ranker {
	// デフォルト値設定
	if config.SortOrder == "" {
		config.SortOrder = "desc"
	}
	if config.Limit == 0 {
		config.Limit = 20
	}
	if config.RandomFactor < 0 {
		config.RandomFactor = 0
	}
	if config.RandomFactor > 1 {
		config.RandomFactor = 1
	}

	return &Ranker{
		Config: config,
	}
}

// Rank マッチにランキング戦略を適用する
func (r *Ranker) Rank(matches []ScoredMatch) []ScoredMatch {
	if len(matches) == 0 {
		return matches
	}

	// オリジナルを変更しないためコピーを作成
	result := make([]ScoredMatch, len(matches))
	copy(result, matches)

	// 1. スコアでソート
	r.sortByScore(result)

	// 2. 多様性を適用（有効な場合）
	if r.Config.Diversity != nil && r.Config.Diversity.Enabled {
		result = r.applyDiversity(result)
	}

	// 3. ランダム性を適用
	if r.Config.RandomFactor > 0 {
		result = r.applyRandomness(result)
	}

	// 4. ページネーション適用
	start := r.Config.Offset
	end := start + r.Config.Limit

	if start >= len(result) {
		result = []ScoredMatch{}
	} else {
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	// 5. ランク付与
	for i := range result {
		result[i].Rank = i + 1
	}

	return result
}

func (r *Ranker) sortByScore(matches []ScoredMatch) {
	sort.Slice(matches, func(i, j int) bool {
		if r.Config.SortOrder == "asc" {
			return matches[i].Score < matches[j].Score
		}
		return matches[i].Score > matches[j].Score
	})
}

func (r *Ranker) applyDiversity(matches []ScoredMatch) []ScoredMatch {
	if r.Config.Diversity == nil || !r.Config.Diversity.Enabled {
		return matches
	}

	groupByField := r.Config.Diversity.GroupByField
	maxPerGroup := r.Config.Diversity.MaxPerGroup

	if groupByField == "" || maxPerGroup <= 0 {
		return matches
	}

	groupCounts := make(map[string]int)
	result := make([]ScoredMatch, 0, len(matches))

	for _, match := range matches {
		// カテゴリカルフィールドからグループ値を抽出
		groupValue := r.extractGroupValue(match.Candidate, groupByField)

		// このグループの上限に達したかチェック
		if groupCounts[groupValue] >= maxPerGroup {
			continue
		}

		result = append(result, match)
		groupCounts[groupValue]++
	}

	return result
}

func (r *Ranker) extractGroupValue(fv *FeatureVector, field string) string {
	// まずカテゴリカルフィールドを試行
	if cat, ok := fv.Categorical[field]; ok {
		// 最大値を持つカテゴリを検索
		maxCategory := ""
		maxValue := 0.0
		for category, value := range cat {
			if value > maxValue {
				maxValue = value
				maxCategory = category
			}
		}
		if maxCategory != "" {
			return maxCategory
		}
	}

	// メタデータを試行
	if meta, ok := fv.Metadata[field]; ok {
		if str, ok := meta.(string); ok {
			return str
		}
	}

	return ""
}

func (r *Ranker) applyRandomness(matches []ScoredMatch) []ScoredMatch {
	if r.Config.RandomFactor <= 0 || len(matches) == 0 {
		return matches
	}

	// スコアにランダム摂動を適用
	perturbed := make([]ScoredMatch, len(matches))
	copy(perturbed, matches)

	for i := range perturbed {
		// [1-randomFactor, 1+randomFactor]の範囲でランダム係数を生成
		randomMultiplier := 1.0 + (rand.Float64()*2-1)*r.Config.RandomFactor

		// ソート用の一時スコアを作成（実際のスコアは変更しない）
		perturbed[i].Score = perturbed[i].Score * randomMultiplier
	}

	// 摂動を加えたスコアで再ソート
	r.sortByScore(perturbed)

	// 注: 摂動を加えたスコアをそのまま保持（ランダム化されたランキングを表すため）
	// 実際のスコア値はランダム摂動を反映するように変更される

	return perturbed
}
