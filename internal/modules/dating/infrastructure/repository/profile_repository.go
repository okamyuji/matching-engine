package repository

import (
	"context"
	"fmt"

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

// FindByUserID ユーザーIDによりプロフィール（タグ・写真を含む）を取得する
func (r *profileRepository) FindByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	row, err := r.q.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	profile := profileFromRow(row)
	if err := loadProfileRelations(ctx, r.q, profile); err != nil {
		return nil, err
	}
	return profile, nil
}

// loadProfileRelations タグと写真を読み込む
func loadProfileRelations(ctx context.Context, q *sqlcgen.Queries, profile *domain.Profile) error {
	tags, err := q.ListProfileTags(ctx, profile.UserID)
	if err != nil {
		return fmt.Errorf("list tags %s: %w", profile.UserID, err)
	}
	profile.Tags = make([]domain.ProfileTag, 0, len(tags))
	for _, t := range tags {
		profile.Tags = append(profile.Tags, domain.ProfileTag{ID: t.ID, UserID: t.UserID, Tag: t.Tag})
	}
	photos, err := q.ListProfilePhotos(ctx, profile.UserID)
	if err != nil {
		return fmt.Errorf("list photos %s: %w", profile.UserID, err)
	}
	profile.Photos = make([]domain.ProfilePhoto, 0, len(photos))
	for _, p := range photos {
		profile.Photos = append(profile.Photos, domain.ProfilePhoto{ID: p.ID, UserID: p.UserID, URL: p.Url, IsPrimary: p.IsPrimary, DisplayOrder: int(p.DisplayOrder), CreatedAt: p.CreatedAt})
	}
	return nil
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
