package matching

import "errors"

var (
	// ErrInvalidID 特徴ベクトルIDが空の場合に返されるエラー
	ErrInvalidID = errors.New("feature vector ID cannot be empty")

	// ErrInvalidType 特徴ベクトルTypeが空の場合に返されるエラー
	ErrInvalidType = errors.New("feature vector type cannot be empty")

	// ErrFieldNotFound 必須フィールドが特徴ベクトルに見つからない場合に返されるエラー
	ErrFieldNotFound = errors.New("required field not found in feature vector")

	// ErrInvalidFieldValue フィールド値が無効な場合に返されるエラー
	ErrInvalidFieldValue = errors.New("invalid field value")

	// ErrEmptyVector 空のベクトルに対して操作を行った場合に返されるエラー
	ErrEmptyVector = errors.New("vector is empty")

	// ErrIncompatibleVectors ベクトルが操作に対して互換性がない場合に返されるエラー
	ErrIncompatibleVectors = errors.New("vectors are incompatible")

	// ErrInvalidConfig 設定が無効な場合に返されるエラー
	ErrInvalidConfig = errors.New("invalid configuration")

	// ErrInvalidWeight 重みが無効な場合に返されるエラー
	ErrInvalidWeight = errors.New("weight must be between 0 and 1")

	// ErrNoComponents スコアリングコンポーネントが定義されていない場合に返されるエラー
	ErrNoComponents = errors.New("no scoring components defined")

	// ErrInvalidTransform 変換タイプが不明な場合に返されるエラー
	ErrInvalidTransform = errors.New("unknown transform type")

	// ErrInvalidFilter フィルタ設定が無効な場合に返されるエラー
	ErrInvalidFilter = errors.New("invalid filter configuration")

	// ErrInvalidOperator フィルタ演算子が不明な場合に返されるエラー
	ErrInvalidOperator = errors.New("unknown filter operator")
)
