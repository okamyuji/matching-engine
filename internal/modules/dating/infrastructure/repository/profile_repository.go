package repository

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/dating/domain"
)

// ProfileRepository プロフィールデータアクセス用インターフェース
type ProfileRepository interface {
	FindByUserID(ctx context.Context, userID string) (*domain.Profile, error)
	Upsert(ctx context.Context, profile *domain.Profile) error
}

// profileRepository ProfileRepositoryのBUN実装
type profileRepository struct {
	db *bun.DB
}

// NewProfileRepository 新しいProfileRepositoryを作成する
func NewProfileRepository(db *bun.DB) ProfileRepository {
	return &profileRepository{db: db}
}

// FindByUserID ユーザーIDによりプロフィールを取得する
func (r *profileRepository) FindByUserID(ctx context.Context, userID string) (*domain.Profile, error) {
	profile := &domain.Profile{}
	err := r.db.NewSelect().
		Model(profile).
		Where("user_id = ?", userID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

// Upsert プロフィールを挿入または更新する
func (r *profileRepository) Upsert(ctx context.Context, profile *domain.Profile) error {
	_, err := r.db.NewInsert().
		Model(profile).
		On("DUPLICATE KEY UPDATE").
		Exec(ctx)
	return err
}
