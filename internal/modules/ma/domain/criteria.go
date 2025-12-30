package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// MAMatchingCriteria M&Aマッチング基準エンティティ
type MAMatchingCriteria struct {
	bun.BaseModel `bun:"table:ma_matching_criteria"`

	CompanyID          string          `bun:"company_id,pk"`
	Purpose            MatchingPurpose `bun:"purpose,notnull"`
	RevenueMin         int64           `bun:"revenue_min"`
	RevenueMax         int64           `bun:"revenue_max"`
	EBITDAMin          int64           `bun:"ebitda_min"`
	EmployeeMin        int             `bun:"employee_min"`
	EmployeeMax        int             `bun:"employee_max"`
	MaxDebtEquityRatio float64         `bun:"max_debt_equity_ratio"`
	UpdatedAt          time.Time       `bun:"updated_at,nullzero,default:current_timestamp"`

	// リレーション
	TargetIndustries []CriteriaIndustry `bun:"rel:has-many,join:company_id=company_id"`
}

// CriteriaIndustry 対象業界
type CriteriaIndustry struct {
	bun.BaseModel `bun:"table:ma_criteria_industries"`

	ID        int64    `bun:"id,pk,autoincrement"`
	CompanyID string   `bun:"company_id,notnull"`
	Industry  Industry `bun:"industry,notnull"`
}

// GetIndustryStrings 業界文字列の配列を取得する
func (c *MAMatchingCriteria) GetIndustryStrings() []string {
	result := make([]string, len(c.TargetIndustries))
	for i, ind := range c.TargetIndustries {
		result[i] = string(ind.Industry)
	}
	return result
}
