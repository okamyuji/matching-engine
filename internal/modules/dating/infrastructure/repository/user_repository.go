package repository

import (
	"context"
	"time"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/dating/domain"
)

// UserRepository ユーザーデータアクセス用インターフェース
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindCandidates(ctx context.Context, userID string, pref *domain.Preference) ([]*UserWithProfile, error)
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
}

// UserWithProfile ユーザーとプロフィールデータを結合する
type UserWithProfile struct {
	User    *domain.User
	Profile *domain.Profile
}

// userRepository UserRepositoryのBUN実装
type userRepository struct {
	db *bun.DB
}

// NewUserRepository 新しいUserRepositoryを作成する
func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{db: db}
}

// FindByID IDによりユーザーを取得する
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	user := &domain.User{}
	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// FindCandidates 設定に基づいて候補ユーザーを取得する
// 要件:
// - 年齢、都道府県、身長、収入、学歴、結婚意向、喫煙、飲酒でフィルタ
// - 認証済みユーザーのみ
// - 最大1000件の候補
func (r *userRepository) FindCandidates(
	ctx context.Context,
	userID string,
	pref *domain.Preference,
) ([]*UserWithProfile, error) {
	var users []*domain.User

	// Profileテーブルと結合してフィルタリング
	query := r.db.NewSelect().
		Model(&users).
		Join("INNER JOIN dating_profiles AS p ON p.user_id = id").
		Where("id != ?", userID).
		Where("verified = ?", true)

	// 年齢フィルタ
	if pref.AgeMin > 0 && pref.AgeMax > 0 {
		minDate := time.Now().AddDate(-pref.AgeMax-1, 0, 0)
		maxDate := time.Now().AddDate(-pref.AgeMin, 0, 0)
		query = query.Where("birth_date BETWEEN ? AND ?", minDate, maxDate)
	}

	// 都道府県フィルタ
	prefStrings := pref.GetPrefectureStrings()
	if len(prefStrings) > 0 {
		query = query.Where("prefecture IN (?)", bun.In(prefStrings))
	}

	// 身長フィルタ
	if pref.HeightMin > 0 {
		query = query.Where("p.height >= ?", pref.HeightMin)
	}
	if pref.HeightMax > 0 {
		query = query.Where("p.height <= ?", pref.HeightMax)
	}

	// 収入フィルタ
	if pref.IncomeMin > 0 {
		query = query.Where("p.income_level >= ?", pref.IncomeMin)
	}

	// 学歴フィルタ
	eduStrings := pref.GetEducationStrings()
	if len(eduStrings) > 0 {
		query = query.Where("p.education IN (?)", bun.In(eduStrings))
	}

	// 結婚意向フィルタ
	marriageStrings := pref.GetMarriageDesireStrings()
	if len(marriageStrings) > 0 {
		query = query.Where("p.marriage_desire IN (?)", bun.In(marriageStrings))
	}

	// 喫煙フィルタ
	smokingStrings := pref.GetSmokingStatusStrings()
	if len(smokingStrings) > 0 {
		query = query.Where("p.smoking IN (?)", bun.In(smokingStrings))
	}

	// 飲酒フィルタ
	drinkingStrings := pref.GetDrinkingStatusStrings()
	if len(drinkingStrings) > 0 {
		query = query.Where("p.drinking IN (?)", bun.In(drinkingStrings))
	}

	query = query.Limit(1000)

	err := query.Scan(ctx)
	if err != nil {
		return nil, err
	}

	// 各ユーザーのプロフィールをロード
	results := make([]*UserWithProfile, 0, len(users))
	for _, user := range users {
		profile := &domain.Profile{}
		err := r.db.NewSelect().
			Model(profile).
			Where("user_id = ?", user.ID).
			Scan(ctx)

		// プロフィールが存在しない場合はnilでOK
		if err != nil && err.Error() != "sql: no rows in result set" {
			return nil, err
		}
		if err != nil {
			profile = nil
		}

		results = append(results, &UserWithProfile{
			User:    user,
			Profile: profile,
		})
	}

	return results, nil
}

// Create 新しいユーザーを挿入する
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.db.NewInsert().
		Model(user).
		Exec(ctx)
	return err
}

// Update 既存のユーザーを更新する
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	_, err := r.db.NewUpdate().
		Model(user).
		Where("id = ?", user.ID).
		Exec(ctx)
	return err
}
