package repository

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/yourorg/matching-engine/internal/modules/ma/domain"
)

// FinancialsRepository 財務情報データアクセス用インターフェース
type FinancialsRepository interface {
	FindByCompanyID(ctx context.Context, companyID string, years int) ([]*domain.Financials, error)
	FindLatest(ctx context.Context, companyID string) (*domain.Financials, error)
	Save(ctx context.Context, financials *domain.Financials) error
}

// financialsRepository FinancialsRepositoryのBUN実装
type financialsRepository struct {
	db *bun.DB
}

// NewFinancialsRepository 新しいFinancialsRepositoryを作成する
func NewFinancialsRepository(db *bun.DB) FinancialsRepository {
	return &financialsRepository{db: db}
}

// FindByCompanyID 企業IDにより指定年数分の財務情報を取得する
func (r *financialsRepository) FindByCompanyID(
	ctx context.Context,
	companyID string,
	years int,
) ([]*domain.Financials, error) {
	var financials []*domain.Financials

	err := r.db.NewSelect().
		Model(&financials).
		Where("company_id = ?", companyID).
		Order("fiscal_year DESC").
		Limit(years).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return financials, nil
}

// FindLatest 企業の最新年度の財務情報を取得する
func (r *financialsRepository) FindLatest(ctx context.Context, companyID string) (*domain.Financials, error) {
	financials := &domain.Financials{}

	err := r.db.NewSelect().
		Model(financials).
		Where("company_id = ?", companyID).
		Order("fiscal_year DESC").
		Limit(1).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return financials, nil
}

// Save 財務情報を保存する（挿入または更新）
func (r *financialsRepository) Save(ctx context.Context, financials *domain.Financials) error {
	_, err := r.db.NewInsert().
		Model(financials).
		On("DUPLICATE KEY UPDATE").
		Set("revenue = VALUES(revenue)").
		Set("ebitda = VALUES(ebitda)").
		Set("net_income = VALUES(net_income)").
		Set("total_assets = VALUES(total_assets)").
		Set("total_liabilities = VALUES(total_liabilities)").
		Set("equity = VALUES(equity)").
		Set("roe = VALUES(roe)").
		Set("roa = VALUES(roa)").
		Set("debt_equity_ratio = VALUES(debt_equity_ratio)").
		Set("current_ratio = VALUES(current_ratio)").
		Exec(ctx)

	return err
}
