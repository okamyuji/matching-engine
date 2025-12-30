package repository

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/ma/domain"
)

// InterestRepository 興味表明データアクセス用インターフェース
type InterestRepository interface {
	Save(ctx context.Context, interest *domain.Interest) error
	FindByToCompany(ctx context.Context, toCompanyID string) ([]*domain.Interest, error)
	Exists(ctx context.Context, fromCompanyID, toCompanyID string) (bool, error)
}

// interestRepository InterestRepositoryのBUN実装
type interestRepository struct {
	db *bun.DB
}

// NewInterestRepository 新しいInterestRepositoryを作成する
func NewInterestRepository(db *bun.DB) InterestRepository {
	return &interestRepository{db: db}
}

// Save 興味表明を保存する
func (r *interestRepository) Save(ctx context.Context, interest *domain.Interest) error {
	_, err := r.db.NewInsert().
		Model(interest).
		Exec(ctx)
	return err
}

// FindByToCompany 企業が受け取った興味表明を取得する
func (r *interestRepository) FindByToCompany(ctx context.Context, toCompanyID string) ([]*domain.Interest, error) {
	var interests []*domain.Interest

	err := r.db.NewSelect().
		Model(&interests).
		Where("to_company_id = ?", toCompanyID).
		Order("created_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return interests, nil
}

// Exists 興味表明が既に存在するかチェックする
func (r *interestRepository) Exists(ctx context.Context, fromCompanyID, toCompanyID string) (bool, error) {
	exists, err := r.db.NewSelect().
		Model((*domain.Interest)(nil)).
		Where("from_company_id = ?", fromCompanyID).
		Where("to_company_id = ?", toCompanyID).
		Exists(ctx)

	return exists, err
}
