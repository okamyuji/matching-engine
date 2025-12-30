package matching

import "math"

// TransformFunc スコア変換関数
// 距離から類似度への変換や非線形変換に使用される
type TransformFunc func(float64) float64

// LinearTransform 線形変換関数を作成する: y = a*x + b
func LinearTransform(a, b float64) TransformFunc {
	return func(x float64) float64 {
		return a*x + b
	}
}

// InverseTransform 逆変換関数を作成する: y = 1 / (1 + x)
// 距離から類似度への変換に有用
func InverseTransform() TransformFunc {
	return func(x float64) float64 {
		return 1.0 / (1.0 + x)
	}
}

// GaussianTransform ガウス変換関数を作成する: y = exp(-(x-μ)² / (2σ²))
// 特定の点付近の値を強調するのに有用
func GaussianTransform(mu, sigma float64) TransformFunc {
	return func(x float64) float64 {
		diff := x - mu
		return math.Exp(-(diff * diff) / (2 * sigma * sigma))
	}
}

// SigmoidTransform シグモイド変換関数を作成する: y = 1 / (1 + exp(-k(x-x0)))
// 滑らかなS字型の変換に有用
func SigmoidTransform(k, x0 float64) TransformFunc {
	return func(x float64) float64 {
		return 1.0 / (1.0 + math.Exp(-k*(x-x0)))
	}
}

// StepTransform ステップ関数を作成する: x >= threshold なら y = 1、そうでなければ 0
// 厳密なカットオフに有用
func StepTransform(threshold float64) TransformFunc {
	return func(x float64) float64 {
		if x >= threshold {
			return 1.0
		}
		return 0.0
	}
}
