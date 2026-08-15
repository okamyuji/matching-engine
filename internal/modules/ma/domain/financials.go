package domain

import "time"

// Financials 財務情報エンティティ
type Financials struct {
	ID               int64
	CompanyID        string
	FiscalYear       int
	Revenue          int64 // 売上高
	EBITDA           int64 // EBITDA
	NetIncome        int64 // 純利益
	TotalAssets      int64 // 総資産
	TotalLiabilities int64
	Equity           int64   // 自己資本
	ROE              float64 // 自己資本利益率
	ROA              float64 // 総資産利益率
	DebtEquityRatio  float64
	CurrentRatio     float64
	CreatedAt        time.Time
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
