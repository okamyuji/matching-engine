package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// Financials 財務情報エンティティ
type Financials struct {
	bun.BaseModel `bun:"table:ma_financials"`

	ID               int64     `bun:"id,pk,autoincrement"`
	CompanyID        string    `bun:"company_id,notnull"`
	FiscalYear       int       `bun:"fiscal_year,notnull"`
	Revenue          int64     `bun:"revenue,notnull"`      // 売上高
	EBITDA           int64     `bun:"ebitda,notnull"`       // EBITDA
	NetIncome        int64     `bun:"net_income,notnull"`   // 純利益
	TotalAssets      int64     `bun:"total_assets,notnull"` // 総資産
	TotalLiabilities int64     `bun:"total_liabilities,notnull"`
	Equity           int64     `bun:"equity,notnull"` // 自己資本
	ROE              float64   `bun:"roe"`            // 自己資本利益率
	ROA              float64   `bun:"roa"`            // 総資産利益率
	DebtEquityRatio  float64   `bun:"debt_equity_ratio"`
	CurrentRatio     float64   `bun:"current_ratio"`
	CreatedAt        time.Time `bun:"created_at,nullzero,default:current_timestamp"`
}

// EBITDAMargin EBITDAマージンを計算する
func (f *Financials) EBITDAMargin() float64 {
	if f.Revenue == 0 {
		return 0
	}
	return float64(f.EBITDA) / float64(f.Revenue)
}

// IsHealthy 財務健全性を判定する
func (f *Financials) IsHealthy() bool {
	return f.ROE > 0 &&
		f.ROA > 0 &&
		f.NetIncome > 0 &&
		f.DebtEquityRatio < 2.0 &&
		f.CurrentRatio > 1.0
}
