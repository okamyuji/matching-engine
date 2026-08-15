package domain

import "time"

// MAMatchingCriteria M&Aマッチング基準エンティティ
type MAMatchingCriteria struct {
	CompanyID          string
	Purpose            MatchingPurpose
	RevenueMin         int64
	RevenueMax         int64
	EBITDAMin          int64
	EmployeeMin        int
	EmployeeMax        int
	MaxDebtEquityRatio float64
	UpdatedAt          time.Time

	// リレーション
	TargetIndustries []CriteriaIndustry
}

// CriteriaIndustry 対象業界
type CriteriaIndustry struct {
	ID        int64
	CompanyID string
	Industry  Industry
}

// GetIndustryStrings 業界文字列の配列を取得する
func (c *MAMatchingCriteria) GetIndustryStrings() []string {
	result := make([]string, len(c.TargetIndustries))
	for i, ind := range c.TargetIndustries {
		result[i] = string(ind.Industry)
	}
	return result
}
