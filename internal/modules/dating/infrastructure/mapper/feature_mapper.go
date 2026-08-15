package mapper

import (
	"time"

	"github.com/okamyuji/matching-engine/internal/core/matching"
	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
)

// DatingFeatureMapper デートドメインモデルを特徴ベクトルに変換する
type DatingFeatureMapper struct{}

// NewDatingFeatureMapper 新しいDatingFeatureMapperを作成する
func NewDatingFeatureMapper() *DatingFeatureMapper {
	return &DatingFeatureMapper{}
}

// ToFeatureVector ユーザーとそのプロフィールを特徴ベクトルに変換する
func (m *DatingFeatureMapper) ToFeatureVector(
	user *domain.User,
	profile *domain.Profile,
) *matching.FeatureVector {
	fv := matching.NewFeatureVector(user.ID, "dating_user")

	// 数値特徴量 (0-1に正規化)
	age := user.Age()
	fv.SetNumerical("age", normalizeAge(age))
	fv.SetNumerical("height", normalizeHeight(profile.Height))
	fv.SetNumerical("income", normalizeIncomeLevel(profile.IncomeLevel))
	fv.SetNumerical("elo", normalizeElo(user.EloRating))
	fv.SetNumerical("activity", normalizeActivity(user.LastActiveAt))
	fv.SetNumerical("recency", normalizeRecency(user.CreatedAt))

	// カテゴリカル特徴量
	fv.SetCategorical("prefecture", string(user.Prefecture), 1.0)
	fv.SetCategorical("body_type", string(profile.BodyType), 1.0)
	fv.SetCategorical("education", string(profile.Education), 1.0)
	fv.SetCategorical("marriage_desire", string(profile.MarriageDesire), 1.0)
	fv.SetCategorical("children_desire", string(profile.ChildrenDesire), 1.0)
	fv.SetCategorical("smoking", string(profile.Smoking), 1.0)
	fv.SetCategorical("drinking", string(profile.Drinking), 1.0)

	// スパース特徴量 (タグ) - リレーションから取得
	for _, tag := range profile.Tags {
		fv.SetSparse("tags", tag.Tag, 1.0)
	}

	// メタデータ (スコアリングには使用しない)
	fv.SetMetadata("gender", user.Gender)
	fv.SetMetadata("nickname", user.Nickname)

	return fv
}

// normalizeAge 年齢を0-1の範囲に正規化する（18-80歳）
func normalizeAge(age int) float64 {
	return matching.NormalizeValue(float64(age), 18, 80)
}

// normalizeHeight 身長を0-1の範囲に正規化する（140-200 cm）
func normalizeHeight(height int) float64 {
	return matching.NormalizeValue(float64(height), 140, 200)
}

// normalizeIncomeLevel 収入レベルを0-1の範囲に正規化する（0-2000万円）
func normalizeIncomeLevel(incomeLevel int) float64 {
	return matching.NormalizeValue(float64(incomeLevel), 0, 2000)
}

// normalizeElo Eloレーティングを0-1の範囲に正規化する（0-2000）
func normalizeElo(elo int) float64 {
	return matching.NormalizeValue(float64(elo), 0, 2000)
}

// normalizeActivity 最終アクティブ時刻に基づいてアクティビティを正規化する
// 1.0 = 現在アクティブ、0.0 = 30日以上前
func normalizeActivity(lastActive time.Time) float64 {
	days := time.Since(lastActive).Hours() / 24
	if days > 30 {
		days = 30
	}
	return 1.0 - (days / 30.0)
}

// normalizeRecency アカウント作成日時に基づいて新規度を正規化する
// 1.0 = 新規作成、0.0 = 1年以上前
func normalizeRecency(createdAt time.Time) float64 {
	days := time.Since(createdAt).Hours() / 24
	if days > 365 {
		days = 365
	}
	return 1.0 - (days / 365.0)
}
