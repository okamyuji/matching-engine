package matching

import (
	"fmt"
	"math"
)

// SimilarityFunction 類似度関数のインターフェース
// 戻り値: 0 (完全不一致) から 1 (完全一致)
type SimilarityFunction interface {
	// Compute 2つの特徴ベクトル間の類似度を計算する
	Compute(a, b *FeatureVector) (float64, error)

	// Name 関数名を返す
	Name() string
}

// CosineSimilarity 埋め込みベクトル間のコサイン類似度を計算する
type CosineSimilarity struct {
	VectorField string
}

// Name "cosine" を返す
func (s *CosineSimilarity) Name() string {
	return "cosine"
}

// Compute コサイン類似度を計算する: cos(θ) = (a·b) / (||a|| ||b||)
func (s *CosineSimilarity) Compute(a, b *FeatureVector) (float64, error) {
	if s.VectorField == "" {
		return 0, fmt.Errorf("%w: cosine similarity requires vector field", ErrInvalidConfig)
	}

	aVec, aOk := a.Embeddings[s.VectorField]
	bVec, bOk := b.Embeddings[s.VectorField]

	if !aOk {
		return 0, fmt.Errorf("%w: vector field '%s' not found in vector A", ErrFieldNotFound, s.VectorField)
	}
	if !bOk {
		return 0, fmt.Errorf("%w: vector field '%s' not found in vector B", ErrFieldNotFound, s.VectorField)
	}

	if len(aVec) == 0 || len(bVec) == 0 {
		return 0, ErrEmptyVector
	}

	if len(aVec) != len(bVec) {
		return 0, fmt.Errorf("%w: vector dimensions do not match (%d vs %d)", ErrIncompatibleVectors, len(aVec), len(bVec))
	}

	// 内積とノルムを計算
	dotProduct := 0.0
	magnitudeA := 0.0
	magnitudeB := 0.0

	for i := range aVec {
		dotProduct += aVec[i] * bVec[i]
		magnitudeA += aVec[i] * aVec[i]
		magnitudeB += bVec[i] * bVec[i]
	}

	magnitudeA = math.Sqrt(magnitudeA)
	magnitudeB = math.Sqrt(magnitudeB)

	// ゼロベクトルの処理
	if magnitudeA == 0 || magnitudeB == 0 {
		return 0, nil
	}

	similarity := dotProduct / (magnitudeA * magnitudeB)

	// 浮動小数点誤差対応のため[-1, 1]にクランプ
	if similarity > 1.0 {
		similarity = 1.0
	}
	if similarity < -1.0 {
		similarity = -1.0
	}

	return similarity, nil
}

// JaccardSimilarity スパース特徴量間のJaccard類似度を計算する
type JaccardSimilarity struct {
	SparseField string
}

// Name "jaccard" を返す
func (s *JaccardSimilarity) Name() string {
	return "jaccard"
}

// Compute Jaccard類似度を計算する: J = |A ∩ B| / |A ∪ B|
func (s *JaccardSimilarity) Compute(a, b *FeatureVector) (float64, error) {
	if s.SparseField == "" {
		return 0, fmt.Errorf("%w: jaccard similarity requires sparse field", ErrInvalidConfig)
	}

	aSet, aOk := a.Sparse[s.SparseField]
	bSet, bOk := b.Sparse[s.SparseField]

	if !aOk {
		return 0, fmt.Errorf("%w: sparse field '%s' not found in vector A", ErrFieldNotFound, s.SparseField)
	}
	if !bOk {
		return 0, fmt.Errorf("%w: sparse field '%s' not found in vector B", ErrFieldNotFound, s.SparseField)
	}

	// 空集合の処理
	if len(aSet) == 0 && len(bSet) == 0 {
		return 1.0, nil // 両方空の場合は同一とみなす
	}
	if len(aSet) == 0 || len(bSet) == 0 {
		return 0.0, nil // 一方が空、もう一方が非空
	}

	// 共通要素を計算
	intersection := 0.0
	for item := range aSet {
		if _, ok := bSet[item]; ok {
			intersection++
		}
	}

	// 和集合を計算
	union := make(map[string]bool)
	for item := range aSet {
		union[item] = true
	}
	for item := range bSet {
		union[item] = true
	}

	if len(union) == 0 {
		return 0.0, nil
	}

	return intersection / float64(len(union)), nil
}

// CategoricalSimilarity カテゴリカル特徴量の一致を判定する
type CategoricalSimilarity struct {
	Field string
}

// Name "categorical" を返す
func (s *CategoricalSimilarity) Name() string {
	return "categorical"
}

// Compute カテゴリが一致する場合は1を、それ以外は0を返す
func (s *CategoricalSimilarity) Compute(a, b *FeatureVector) (float64, error) {
	if s.Field == "" {
		return 0, fmt.Errorf("%w: categorical similarity requires field", ErrInvalidConfig)
	}

	aCat, aOk := a.Categorical[s.Field]
	bCat, bOk := b.Categorical[s.Field]

	if !aOk {
		return 0, fmt.Errorf("%w: categorical field '%s' not found in vector A", ErrFieldNotFound, s.Field)
	}
	if !bOk {
		return 0, fmt.Errorf("%w: categorical field '%s' not found in vector B", ErrFieldNotFound, s.Field)
	}

	// 各ベクトルで最大値を持つカテゴリを検索
	aCategory := ""
	aMaxValue := 0.0
	for cat, val := range aCat {
		if val > aMaxValue {
			aMaxValue = val
			aCategory = cat
		}
	}

	bCategory := ""
	bMaxValue := 0.0
	for cat, val := range bCat {
		if val > bMaxValue {
			bMaxValue = val
			bCategory = cat
		}
	}

	if aCategory == bCategory && aCategory != "" {
		return 1.0, nil
	}

	return 0.0, nil
}
