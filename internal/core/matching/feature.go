package matching

import (
	"maps"
	"math"
)

// FeatureVector 任意のエンティティを数値表現として表す
type FeatureVector struct {
	// 一意識別子
	ID string `json:"id"`

	// エンティティタイプ（例: "dating_user", "ma_company"）
	Type string `json:"type"`

	// 数値特徴（0-1に正規化）
	// 例: {"age": 0.35, "height": 0.68}
	Numerical map[string]float64 `json:"numerical"`

	// カテゴリ特徴（ワンホットエンコーディング）
	// 例: {"prefecture": {"tokyo": 1.0}, "industry": {"tech": 1.0}}
	Categorical map[string]map[string]float64 `json:"categorical"`

	// 埋め込みベクトル（密ベクトル）
	// 例: {"text_embedding": [0.1, 0.2, ...]}
	Embeddings map[string][]float64 `json:"embeddings"`

	// スパース特徴（タグ、キーワード）
	// 例: {"tags": {"sports": 1.0, "travel": 0.8}}
	Sparse map[string]map[string]float64 `json:"sparse"`

	// 時系列統計
	// 例: {"revenue_trend": {...}}
	TimeSeries map[string]*TimeSeriesStats `json:"time_series,omitempty"`

	// メタデータ（スコアリングには使用しない）
	Metadata map[string]any `json:"metadata,omitempty"`
}

// TimeSeriesStats 時系列データの統計的測度を表す
type TimeSeriesStats struct {
	Mean       float64 `json:"mean"`       // 平均値
	Std        float64 `json:"std"`        // 標準偏差
	Min        float64 `json:"min"`        // 最小値
	Max        float64 `json:"max"`        // 最大値
	Trend      float64 `json:"trend"`      // 線形トレンド係数
	Volatility float64 `json:"volatility"` // ボラティリティ
}

// NewFeatureVector 指定されたIDとタイプで新しいFeatureVectorを作成する
func NewFeatureVector(id, entityType string) *FeatureVector {
	return &FeatureVector{
		ID:          id,
		Type:        entityType,
		Numerical:   make(map[string]float64),
		Categorical: make(map[string]map[string]float64),
		Embeddings:  make(map[string][]float64),
		Sparse:      make(map[string]map[string]float64),
		TimeSeries:  make(map[string]*TimeSeriesStats),
		Metadata:    make(map[string]any),
	}
}

// SetNumerical 数値特徴を設定する
func (f *FeatureVector) SetNumerical(key string, value float64) {
	if f.Numerical == nil {
		f.Numerical = make(map[string]float64)
	}
	f.Numerical[key] = value
}

// SetCategorical カテゴリ特徴を設定する
func (f *FeatureVector) SetCategorical(key, category string, value float64) {
	if f.Categorical == nil {
		f.Categorical = make(map[string]map[string]float64)
	}
	if f.Categorical[key] == nil {
		f.Categorical[key] = make(map[string]float64)
	}
	f.Categorical[key][category] = value
}

// SetEmbedding 埋め込みベクトルを設定する
func (f *FeatureVector) SetEmbedding(key string, vector []float64) {
	if f.Embeddings == nil {
		f.Embeddings = make(map[string][]float64)
	}
	// 外部からの変更を防ぐためコピーを作成
	copied := make([]float64, len(vector))
	copy(copied, vector)
	f.Embeddings[key] = copied
}

// SetSparse スパース特徴を設定する
func (f *FeatureVector) SetSparse(key, item string, weight float64) {
	if f.Sparse == nil {
		f.Sparse = make(map[string]map[string]float64)
	}
	if f.Sparse[key] == nil {
		f.Sparse[key] = make(map[string]float64)
	}
	f.Sparse[key][item] = weight
}

// EnsureSparse スパース特徴のキーを空集合として確保する。要素が無い場合でも
// Jaccard 類似度がフィールド未定義エラーにならないようにするために使う
func (f *FeatureVector) EnsureSparse(key string) {
	if f.Sparse == nil {
		f.Sparse = make(map[string]map[string]float64)
	}
	if f.Sparse[key] == nil {
		f.Sparse[key] = make(map[string]float64)
	}
}

// SetTimeSeries 時系列統計を設定する
func (f *FeatureVector) SetTimeSeries(key string, stats *TimeSeriesStats) {
	if f.TimeSeries == nil {
		f.TimeSeries = make(map[string]*TimeSeriesStats)
	}
	f.TimeSeries[key] = stats
}

// SetMetadata メタデータを設定する
func (f *FeatureVector) SetMetadata(key string, value any) {
	if f.Metadata == nil {
		f.Metadata = make(map[string]any)
	}
	f.Metadata[key] = value
}

// Validate 特徴ベクトルを検証する
func (f *FeatureVector) Validate() error {
	if f.ID == "" {
		return ErrInvalidID
	}
	if f.Type == "" {
		return ErrInvalidType
	}
	return nil
}

// Clone 特徴ベクトルのディープコピーを作成する
func (f *FeatureVector) Clone() *FeatureVector {
	clone := &FeatureVector{
		ID:          f.ID,
		Type:        f.Type,
		Numerical:   make(map[string]float64),
		Categorical: make(map[string]map[string]float64),
		Embeddings:  make(map[string][]float64),
		Sparse:      make(map[string]map[string]float64),
		TimeSeries:  make(map[string]*TimeSeriesStats),
		Metadata:    make(map[string]any),
	}

	// 数値をコピー
	maps.Copy(clone.Numerical, f.Numerical)

	// カテゴリをコピー
	for k, cats := range f.Categorical {
		clone.Categorical[k] = make(map[string]float64)
		maps.Copy(clone.Categorical[k], cats)
	}

	// 埋め込みをコピー
	for k, vec := range f.Embeddings {
		copied := make([]float64, len(vec))
		copy(copied, vec)
		clone.Embeddings[k] = copied
	}

	// スパースをコピー
	for k, items := range f.Sparse {
		clone.Sparse[k] = make(map[string]float64)
		maps.Copy(clone.Sparse[k], items)
	}

	// 時系列をコピー
	for k, stats := range f.TimeSeries {
		clone.TimeSeries[k] = &TimeSeriesStats{
			Mean:       stats.Mean,
			Std:        stats.Std,
			Min:        stats.Min,
			Max:        stats.Max,
			Trend:      stats.Trend,
			Volatility: stats.Volatility,
		}
	}

	// メタデータをコピー（シャローコピー）
	maps.Copy(clone.Metadata, f.Metadata)

	return clone
}

// NormalizeValue 値を0-1の範囲に正規化する
func NormalizeValue(value, min, max float64) float64 {
	if max == min {
		return 0.5
	}
	normalized := (value - min) / (max - min)
	// [0, 1]にクランプ
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

// ComputeTimeSeriesStats 時系列データから統計的測度を計算する
func ComputeTimeSeriesStats(values []float64) *TimeSeriesStats {
	if len(values) == 0 {
		return &TimeSeriesStats{}
	}

	// 平均を計算
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// 標準偏差を計算
	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	std := math.Sqrt(variance / float64(len(values)))

	// 最小値と最大値を見つける
	min := values[0]
	max := values[0]
	for _, v := range values[1:] {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
	}

	// 最小二乗法を使用して線形トレンドを計算
	trend := 0.0
	if len(values) > 1 {
		n := float64(len(values))
		sumX := 0.0
		sumY := 0.0
		sumXY := 0.0
		sumXX := 0.0

		for i, y := range values {
			x := float64(i)
			sumX += x
			sumY += y
			sumXY += x * y
			sumXX += x * x
		}

		trend = (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	}

	// ボラティリティを計算（リターンの標準偏差）
	volatility := 0.0
	if len(values) > 1 {
		returns := make([]float64, len(values)-1)
		for i := 1; i < len(values); i++ {
			if values[i-1] != 0 {
				returns[i-1] = (values[i] - values[i-1]) / values[i-1]
			}
		}

		returnMean := 0.0
		for _, r := range returns {
			returnMean += r
		}
		returnMean /= float64(len(returns))

		returnVariance := 0.0
		for _, r := range returns {
			diff := r - returnMean
			returnVariance += diff * diff
		}
		volatility = math.Sqrt(returnVariance / float64(len(returns)))
	}

	return &TimeSeriesStats{
		Mean:       mean,
		Std:        std,
		Min:        min,
		Max:        max,
		Trend:      trend,
		Volatility: volatility,
	}
}
