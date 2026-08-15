package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository/sqlcgen"
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

// userRepository UserRepository の sqlc 実装
type userRepository struct {
	db DB
	q  *sqlcgen.Queries
}

// NewUserRepository 新しい UserRepository を作成する
func NewUserRepository(db DB) UserRepository {
	return &userRepository{db: db, q: sqlcgen.New(db)}
}

// FindByID IDによりユーザーを取得する
func (r *userRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row, err := r.q.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return userFromRow(row), nil
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
	params := sqlcgen.FindCandidateUsersParams{
		UserID:           userID,
		Prefectures:      pref.GetPrefectureStrings(),
		Educations:       pref.GetEducationStrings(),
		MarriageDesires:  pref.GetMarriageDesireStrings(),
		SmokingStatuses:  pref.GetSmokingStatusStrings(),
		DrinkingStatuses: pref.GetDrinkingStatusStrings(),
	}
	// 年齢は生年月日の範囲に変換する。AgeMax 歳の人は誕生日の前日まで含めるため 1 年広く取る
	if pref.AgeMin > 0 && pref.AgeMax > 0 {
		now := time.Now()
		minDate := now.AddDate(-pref.AgeMax-1, 0, 0)
		maxDate := now.AddDate(-pref.AgeMin, 0, 0)
		params.BirthMin = &minDate
		params.BirthMax = &maxDate
	}
	if pref.HeightMin > 0 {
		params.HeightMin = int32PtrFromInt(pref.HeightMin)
	}
	if pref.HeightMax > 0 {
		params.HeightMax = int32PtrFromInt(pref.HeightMax)
	}
	if pref.IncomeMin > 0 {
		params.IncomeMin = int32PtrFromInt(pref.IncomeMin)
	}

	rows, err := r.q.FindCandidateUsers(ctx, params)
	if err != nil {
		return nil, err
	}

	results := make([]*UserWithProfile, 0, len(rows))
	for _, row := range rows {
		user := userFromRow(row)
		profile, err := loadProfile(ctx, r.q, user.ID)
		if err != nil {
			return nil, err
		}
		results = append(results, &UserWithProfile{User: user, Profile: profile})
	}
	return results, nil
}

// Create 新しいユーザーを挿入する
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	now := time.Now()
	createdAt := user.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	lastActive := user.LastActiveAt
	if lastActive.IsZero() {
		lastActive = now
	}
	return r.q.CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:           user.ID,
		Nickname:     user.Nickname,
		Gender:       string(user.Gender),
		BirthDate:    user.BirthDate,
		Prefecture:   string(user.Prefecture),
		Verified:     user.Verified,
		EloRating:    int32(user.EloRating), //nolint:gosec // レーティングは小さな整数
		CreatedAt:    createdAt,
		LastActiveAt: lastActive,
	})
}

// Update 既存のユーザーを更新する
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	lastActive := user.LastActiveAt
	if lastActive.IsZero() {
		lastActive = time.Now()
	}
	return r.q.UpdateUser(ctx, sqlcgen.UpdateUserParams{
		ID:           user.ID,
		Nickname:     user.Nickname,
		Gender:       string(user.Gender),
		BirthDate:    user.BirthDate,
		Prefecture:   string(user.Prefecture),
		Verified:     user.Verified,
		EloRating:    int32(user.EloRating), //nolint:gosec // レーティングは小さな整数
		LastActiveAt: lastActive,
	})
}

func userFromRow(row sqlcgen.DatingUser) *domain.User {
	return &domain.User{
		ID:           row.ID,
		Nickname:     row.Nickname,
		Gender:       domain.Gender(row.Gender),
		BirthDate:    row.BirthDate,
		Prefecture:   domain.Prefecture(row.Prefecture),
		Verified:     row.Verified,
		EloRating:    int(row.EloRating),
		CreatedAt:    row.CreatedAt,
		LastActiveAt: row.LastActiveAt,
	}
}

// loadProfile プロフィールを取得する。存在しなければ nil を返す
func loadProfile(ctx context.Context, q *sqlcgen.Queries, userID string) (*domain.Profile, error) {
	row, err := q.GetProfile(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil //nolint:nilnil // プロフィール未設定は正常系
		}
		return nil, fmt.Errorf("load profile %s: %w", userID, err)
	}
	return profileFromRow(row), nil
}
