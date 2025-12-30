package repository

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/ma/domain"
)

// MAMatchRepository M&Aマッチデータアクセス用インターフェース
type MAMatchRepository interface {
	Save(ctx context.Context, match *domain.MAMatch) error
	FindMutual(ctx context.Context, companyID string) ([]*domain.MAMatch, error)
	FindByID(ctx context.Context, matchID string) (*domain.MAMatch, error)
}

// maMatchRepository MAMatchRepositoryのBUN実装
type maMatchRepository struct {
	db *bun.DB
}

// NewMAMatchRepository 新しいMAMatchRepositoryを作成する
func NewMAMatchRepository(db *bun.DB) MAMatchRepository {
	return &maMatchRepository{db: db}
}

// Save マッチを保存する
func (r *maMatchRepository) Save(ctx context.Context, match *domain.MAMatch) error {
	_, err := r.db.NewInsert().
		Model(match).
		Exec(ctx)
	return err
}

// FindMutual 企業の相互マッチを取得する（双方向の興味表明がある）
func (r *maMatchRepository) FindMutual(ctx context.Context, companyID string) ([]*domain.MAMatch, error) {
	var matches []*domain.MAMatch

	// 双方向の興味表明が存在するマッチのみを取得
	err := r.db.NewSelect().
		Model(&matches).
		Where("company_id_a = ? OR company_id_b = ?", companyID, companyID).
		Where("EXISTS (SELECT 1 FROM ma_interests WHERE from_company_id = company_id_a AND to_company_id = company_id_b)").
		Where("EXISTS (SELECT 1 FROM ma_interests WHERE from_company_id = company_id_b AND to_company_id = company_id_a)").
		Order("matched_at DESC").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return matches, nil
}

// FindByID IDによりマッチを取得する
func (r *maMatchRepository) FindByID(ctx context.Context, matchID string) (*domain.MAMatch, error) {
	match := &domain.MAMatch{}

	err := r.db.NewSelect().
		Model(match).
		Where("id = ?", matchID).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return match, nil
}
