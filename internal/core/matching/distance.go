package matching

import (
	"fmt"
	"math"
)

// DistanceFunction 距離関数のインターフェース
// 戻り値: 0 (完全一致) から ∞ (完全不一致)
type DistanceFunction interface {
	// Compute 2つの特徴ベクトル間の距離を計算する
	Compute(a, b *FeatureVector) (float64, error)

	// Name 関数名を返す
	Name() string
}

// EuclideanDistance 数値特徴量間のユークリッド距離を計算する
type EuclideanDistance struct {
	Fields []string
}

// Name "euclidean" を返す
func (d *EuclideanDistance) Name() string {
	return "euclidean"
}

// Compute ユークリッド距離を計算する: d = √Σ(xi - yi)²
func (d *EuclideanDistance) Compute(a, b *FeatureVector) (float64, error) {
	if len(d.Fields) == 0 {
		return 0, fmt.Errorf("%w: euclidean distance requires at least one field", ErrInvalidConfig)
	}

	sumSquares := 0.0
	for _, field := range d.Fields {
		aVal, aOk := a.Numerical[field]
		bVal, bOk := b.Numerical[field]

		if !aOk {
			return 0, fmt.Errorf("%w: field '%s' not found in vector A", ErrFieldNotFound, field)
		}
		if !bOk {
			return 0, fmt.Errorf("%w: field '%s' not found in vector B", ErrFieldNotFound, field)
		}

		diff := aVal - bVal
		sumSquares += diff * diff
	}

	return math.Sqrt(sumSquares), nil
}

// ManhattanDistance 数値特徴量間のマンハッタン距離を計算する
type ManhattanDistance struct {
	Fields []string
}

// Name "manhattan" を返す
func (d *ManhattanDistance) Name() string {
	return "manhattan"
}

// Compute マンハッタン距離を計算する: d = Σ|xi - yi|
func (d *ManhattanDistance) Compute(a, b *FeatureVector) (float64, error) {
	if len(d.Fields) == 0 {
		return 0, fmt.Errorf("%w: manhattan distance requires at least one field", ErrInvalidConfig)
	}

	sumAbs := 0.0
	for _, field := range d.Fields {
		aVal, aOk := a.Numerical[field]
		bVal, bOk := b.Numerical[field]

		if !aOk {
			return 0, fmt.Errorf("%w: field '%s' not found in vector A", ErrFieldNotFound, field)
		}
		if !bOk {
			return 0, fmt.Errorf("%w: field '%s' not found in vector B", ErrFieldNotFound, field)
		}

		sumAbs += math.Abs(aVal - bVal)
	}

	return sumAbs, nil
}
