package repository

import (
	"context"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
	"github.com/okamyuji/matching-engine/internal/modules/dating/infrastructure/repository/sqlcgen"
)

// ProfileRepository プロフィールデータアクセス用インターフェース
type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	Upsert(ctx context.Context, profile *domain.Profile) error
}

// profileRepository ProfileRepository の sqlc 実装
type profileRepository struct {
	q *sqlcgen.Queries
}

// NewProfileRepository 新しい ProfileRepository を作成する
func NewProfileRepository(db DB) ProfileRepository {
	return &profileRepository{q: sqlcgen.New(db)}
}

// FindByUserID ユーザーIDによりプロフィールを取得する
func (r *profileRepository) FindByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	row, err := r.q.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	return profileFromRow(row), nil
}

// Upsert プロフィールを挿入または更新する
func (r *profileRepository) Upsert(ctx context.Context, profile *domain.Profile) error {
	return r.q.UpsertProfile(ctx, sqlcgen.UpsertProfileParams{
		UserID:           profile.UserID,
		Height:           int32PtrFromInt(profile.Height),
		BodyType:         strPtr(string(profile.BodyType)),
		Education:        strPtr(string(profile.Education)),
		Occupation:       strPtr(profile.Occupation),
		IncomeLevel:      int32PtrFromInt(profile.IncomeLevel),
		MarriageDesire:   strPtr(string(profile.MarriageDesire)),
		ChildrenDesire:   strPtr(string(profile.ChildrenDesire)),
		Smoking:          strPtr(string(profile.Smoking)),
		Drinking:         strPtr(string(profile.Drinking)),
		SelfIntroduction: strPtr(profile.SelfIntroduction),
	})
}

func profileFromRow(row sqlcgen.DatingProfile) *domain.Profile {
	return &domain.Profile{
		UserID:           row.UserID,
		Height:           intFromInt32Ptr(row.Height),
		BodyType:         domain.BodyType(strFromPtr(row.BodyType)),
		Education:        domain.Education(strFromPtr(row.Education)),
		Occupation:       strFromPtr(row.Occupation),
		IncomeLevel:      intFromInt32Ptr(row.IncomeLevel),
		MarriageDesire:   domain.MarriageDesire(strFromPtr(row.MarriageDesire)),
		ChildrenDesire:   domain.ChildrenDesire(strFromPtr(row.ChildrenDesire)),
		Smoking:          domain.SmokingStatus(strFromPtr(row.Smoking)),
		Drinking:         domain.DrinkingStatus(strFromPtr(row.Drinking)),
		SelfIntroduction: strFromPtr(row.SelfIntroduction),
		UpdatedAt:        row.UpdatedAt,
	}
}
