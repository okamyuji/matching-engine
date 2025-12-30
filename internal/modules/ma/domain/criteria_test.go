package domain

import (
	"testing"
)

func TestMAMatchingCriteria_Validation(t *testing.T) {
	criteria := &MAMatchingCriteria{
		CompanyID:          "company-123",
		Purpose:            PurposeBuyer,
		EmployeeMin:        50,
		EmployeeMax:        500,
		RevenueMin:         10000000,
		RevenueMax:         100000000,
		EBITDAMin:          1000000,
		MaxDebtEquityRatio: 2.0,
	}

	if criteria.CompanyID != "company-123" {
		t.Errorf("CompanyID = %v, want %v", criteria.CompanyID, "company-123")
	}
	if criteria.Purpose != PurposeBuyer {
		t.Errorf("Purpose = %v, want %v", criteria.Purpose, PurposeBuyer)
	}
	if criteria.EmployeeMin != 50 {
		t.Errorf("EmployeeMin = %v, want %v", criteria.EmployeeMin, 50)
	}
	if criteria.EmployeeMax != 500 {
		t.Errorf("EmployeeMax = %v, want %v", criteria.EmployeeMax, 500)
	}
	if criteria.RevenueMin != 10000000 {
		t.Errorf("RevenueMin = %v, want %v", criteria.RevenueMin, 10000000)
	}
	if criteria.RevenueMax != 100000000 {
		t.Errorf("RevenueMax = %v, want %v", criteria.RevenueMax, 100000000)
	}
	if criteria.EBITDAMin != 1000000 {
		t.Errorf("EBITDAMin = %v, want %v", criteria.EBITDAMin, 1000000)
	}
	if criteria.MaxDebtEquityRatio != 2.0 {
		t.Errorf("MaxDebtEquityRatio = %v, want %v", criteria.MaxDebtEquityRatio, 2.0)
	}
}

func TestMAMatchingCriteria_TargetIndustries(t *testing.T) {
	criteria := &MAMatchingCriteria{
		TargetIndustries: []CriteriaIndustry{
			{CompanyID: "company-123", Industry: IndustryTechnology},
			{CompanyID: "company-123", Industry: IndustryFinance},
			{CompanyID: "company-123", Industry: IndustryHealthcare},
		},
	}

	if len(criteria.TargetIndustries) != 3 {
		t.Errorf("len(TargetIndustries) = %v, want %v", len(criteria.TargetIndustries), 3)
	}

	// 業界タイプの確認
	expectedIndustries := []Industry{IndustryTechnology, IndustryFinance, IndustryHealthcare}
	for i, ti := range criteria.TargetIndustries {
		if ti.Industry != expectedIndustries[i] {
			t.Errorf("TargetIndustries[%d].Industry = %v, want %v", i, ti.Industry, expectedIndustries[i])
		}
		if ti.CompanyID != "company-123" {
			t.Errorf("TargetIndustries[%d].CompanyID = %v, want %v", i, ti.CompanyID, "company-123")
		}
	}
}

func TestMAMatchingCriteria_GetIndustryStrings(t *testing.T) {
	criteria := &MAMatchingCriteria{
		CompanyID: "company-123",
		Purpose:   PurposeBuyer,
		TargetIndustries: []CriteriaIndustry{
			{CompanyID: "company-123", Industry: IndustryTechnology},
			{CompanyID: "company-123", Industry: IndustryFinance},
			{CompanyID: "company-123", Industry: IndustryHealthcare},
		},
	}

	industryStrings := criteria.GetIndustryStrings()

	if len(industryStrings) != 3 {
		t.Errorf("len(GetIndustryStrings()) = %v, want %v", len(industryStrings), 3)
	}

	expectedStrings := []string{"technology", "finance", "healthcare"}
	for i, str := range industryStrings {
		if str != expectedStrings[i] {
			t.Errorf("industryStrings[%d] = %v, want %v", i, str, expectedStrings[i])
		}
	}
}

func TestCriteriaIndustry(t *testing.T) {
	ci := CriteriaIndustry{
		ID:        1,
		CompanyID: "company-123",
		Industry:  IndustryTechnology,
	}

	if ci.ID != 1 {
		t.Errorf("ID = %v, want %v", ci.ID, 1)
	}
	if ci.CompanyID != "company-123" {
		t.Errorf("CompanyID = %v, want %v", ci.CompanyID, "company-123")
	}
	if ci.Industry != IndustryTechnology {
		t.Errorf("Industry = %v, want %v", ci.Industry, IndustryTechnology)
	}
}

func TestMAMatchingCriteria_MinMaxValidation(t *testing.T) {
	tests := []struct {
		name        string
		criteria    *MAMatchingCriteria
		description string
	}{
		{
			name: "ゼロ値の境界条件",
			criteria: &MAMatchingCriteria{
				CompanyID:   "company-123",
				Purpose:     PurposeBuyer,
				EmployeeMin: 0,
				EmployeeMax: 0,
				RevenueMin:  0,
				RevenueMax:  0,
			},
			description: "最小値・最大値がゼロ",
		},
		{
			name: "大きな値の範囲",
			criteria: &MAMatchingCriteria{
				CompanyID:   "company-123",
				Purpose:     PurposeSeller,
				EmployeeMin: 10000,
				EmployeeMax: 50000,
				RevenueMin:  1000000000,
				RevenueMax:  10000000000,
			},
			description: "大企業向けの範囲",
		},
		{
			name: "最小値のみ設定",
			criteria: &MAMatchingCriteria{
				CompanyID:   "company-123",
				Purpose:     PurposeBuyer,
				EmployeeMin: 100,
				RevenueMin:  5000000,
			},
			description: "最小値のみで最大値は未設定",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.criteria.CompanyID != "company-123" {
				t.Errorf("CompanyID = %v, want %v (case: %s)", tt.criteria.CompanyID, "company-123", tt.description)
			}
			// 構造体が正しく作成できることを確認
			if tt.criteria.Purpose != PurposeBuyer && tt.criteria.Purpose != PurposeSeller {
				t.Errorf("Purpose = %v, want PurposeBuyer or PurposeSeller (case: %s)", tt.criteria.Purpose, tt.description)
			}
		})
	}
}
