package repository

import (
	"context"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository/sqlcgen"
)

// FinancialsRepository 財務データアクセス用インターフェース
type FinancialsRepository interface {
	FindByCompanyID(ctx context.Context, companyID string, years int) ([]*domain.Financials, error)
	FindLatest(ctx context.Context, companyID string) (*domain.Financials, error)
	Save(ctx context.Context, financials *domain.Financials) error
}

// financialsRepository FinancialsRepository の sqlc 実装
type financialsRepository struct {
	q *sqlcgen.Queries
}

// NewFinancialsRepository 新しい FinancialsRepository を作成する
func NewFinancialsRepository(db DB) FinancialsRepository {
	return &financialsRepository{q: sqlcgen.New(db)}
}

// FindByCompanyID 企業の財務データを新しい年度から years 件取得する
func (r *financialsRepository) FindByCompanyID(ctx context.Context, companyID string, years int) ([]*domain.Financials, error) {
	rows, err := r.q.ListFinancialsByCompany(ctx, sqlcgen.ListFinancialsByCompanyParams{
		CompanyID: companyID,
		Limit:     int32(years), //nolint:gosec // 取得年数
	})
	if err != nil {
		return nil, err
	}
	return financialsFromRows(rows), nil
}

// FindLatest 企業の最新年度の財務データを取得する
func (r *financialsRepository) FindLatest(ctx context.Context, companyID string) (*domain.Financials, error) {
	row, err := r.q.GetLatestFinancials(ctx, companyID)
	if err != nil {
		return nil, err
	}
	return financialsFromRow(row), nil
}

// Save 財務データを挿入または（同一企業・同一年度なら）更新する
func (r *financialsRepository) Save(ctx context.Context, f *domain.Financials) error {
	createdAt := f.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	id, err := r.q.UpsertFinancials(ctx, sqlcgen.UpsertFinancialsParams{
		CompanyID:        f.CompanyID,
		FiscalYear:       int32(f.FiscalYear), //nolint:gosec // 年度
		Revenue:          f.Revenue,
		Ebitda:           f.EBITDA,
		NetIncome:        f.NetIncome,
		TotalAssets:      f.TotalAssets,
		TotalLiabilities: f.TotalLiabilities,
		Equity:           f.Equity,
		Roe:              float64Ptr(f.ROE),
		Roa:              float64Ptr(f.ROA),
		DebtEquityRatio:  float64Ptr(f.DebtEquityRatio),
		CurrentRatio:     float64Ptr(f.CurrentRatio),
		CreatedAt:        createdAt,
	})
	if err != nil {
		return err
	}
	f.ID = id
	return nil
}

func financialsFromRow(row sqlcgen.MaFinancial) *domain.Financials {
	return &domain.Financials{
		ID:               row.ID,
		CompanyID:        row.CompanyID,
		FiscalYear:       int(row.FiscalYear),
		Revenue:          row.Revenue,
		EBITDA:           row.Ebitda,
		NetIncome:        row.NetIncome,
		TotalAssets:      row.TotalAssets,
		TotalLiabilities: row.TotalLiabilities,
		Equity:           row.Equity,
		ROE:              float64FromPtr(row.Roe),
		ROA:              float64FromPtr(row.Roa),
		DebtEquityRatio:  float64FromPtr(row.DebtEquityRatio),
		CurrentRatio:     float64FromPtr(row.CurrentRatio),
		CreatedAt:        row.CreatedAt,
	}
}

func financialsFromRows(rows []sqlcgen.MaFinancial) []*domain.Financials {
	out := make([]*domain.Financials, 0, len(rows))
	for _, row := range rows {
		out = append(out, financialsFromRow(row))
	}
	return out
}
