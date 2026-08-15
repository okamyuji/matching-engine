package mapper

import (
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
)

func TestMAFeatureMapper_ToFeatureVector(t *testing.T) {
	mapper := NewMAFeatureMapper()

	now := time.Now()
	company := &domain.Company{
		ID:              "company-123",
		Name:            "Test Corp",
		Industry:        domain.IndustryTechnology,
		Location:        "JP",
		EmployeeCount:   500,
		Founded:         now.AddDate(-10, 0, 0),
		ListingStatus:   domain.ListingPublic,
		MatchingPurpose: domain.PurposeBuyer,
		Technologies: []*domain.CompanyTechnology{
			{CompanyID: "company-123", Technology: "AI"},
			{CompanyID: "company-123", Technology: "Cloud"},
		},
		Markets: []*domain.CompanyMarket{
			{CompanyID: "company-123", Market: "B2B"},
			{CompanyID: "company-123", Market: "Enterprise"},
		},
	}

	financials := []*domain.Financials{
		{
			ID:               1,
			CompanyID:        "company-123",
			FiscalYear:       2024,
			Revenue:          10000000000, // 100億円
			EBITDA:           2000000000,  // 20億円
			NetIncome:        1500000000,
			TotalAssets:      50000000000,
			TotalLiabilities: 30000000000,
			Equity:           20000000000,
			ROE:              7.5,
			ROA:              3.0,
			DebtEquityRatio:  1.5,
			CurrentRatio:     2.0,
		},
		{
			ID:         2,
			CompanyID:  "company-123",
			FiscalYear: 2023,
			Revenue:    9000000000,
			EBITDA:     1800000000,
		},
		{
			ID:         3,
			CompanyID:  "company-123",
			FiscalYear: 2022,
			Revenue:    8000000000,
			EBITDA:     1600000000,
		},
	}

	criteria := &domain.MAMatchingCriteria{
		CompanyID: "company-123",
		Purpose:   domain.PurposeBuyer,
	}

	fv := mapper.ToFeatureVector(company, financials, criteria)

	// 基本プロパティの確認
	if fv.ID != "company-123" {
		t.Errorf("ID = %v, want company-123", fv.ID)
	}
	if fv.Type != "ma_company" {
		t.Errorf("Type = %v, want ma_company", fv.Type)
	}

	// 数値特徴の確認
	if _, ok := fv.Numerical["revenue"]; !ok {
		t.Error("revenue feature not found")
	}
	if _, ok := fv.Numerical["ebitda_margin"]; !ok {
		t.Error("ebitda_margin feature not found")
	}
	if _, ok := fv.Numerical["roe"]; !ok {
		t.Error("roe feature not found")
	}
	if _, ok := fv.Numerical["roa"]; !ok {
		t.Error("roa feature not found")
	}
	if _, ok := fv.Numerical["debt_equity_ratio"]; !ok {
		t.Error("debt_equity_ratio feature not found")
	}
	if _, ok := fv.Numerical["current_ratio"]; !ok {
		t.Error("current_ratio feature not found")
	}
	if _, ok := fv.Numerical["employee_count"]; !ok {
		t.Error("employee_count feature not found")
	}

	// カテゴリ特徴の確認
	if fv.Categorical["industry"]["technology"] != 1.0 {
		t.Error("industry categorical feature incorrect")
	}
	if fv.Categorical["location"]["JP"] != 1.0 {
		t.Error("location categorical feature incorrect")
	}
	if fv.Categorical["listing_status"]["public"] != 1.0 {
		t.Error("listing_status categorical feature incorrect")
	}
	if _, ok := fv.Categorical["stage"]; !ok {
		t.Error("stage categorical feature not found")
	}

	// スパース特徴の確認
	if fv.Sparse["technology"]["AI"] != 1.0 {
		t.Error("AI technology not found")
	}
	if fv.Sparse["technology"]["Cloud"] != 1.0 {
		t.Error("Cloud technology not found")
	}
	if fv.Sparse["market"]["B2B"] != 1.0 {
		t.Error("B2B market not found")
	}
	if fv.Sparse["market"]["Enterprise"] != 1.0 {
		t.Error("Enterprise market not found")
	}

	// 時系列特徴の確認（3年以上のデータがある場合）
	if _, ok := fv.TimeSeries["revenue_trend"]; !ok {
		t.Error("revenue_trend time series feature not found")
	}
	if _, ok := fv.TimeSeries["ebitda_trend"]; !ok {
		t.Error("ebitda_trend time series feature not found")
	}

	// メタデータの確認
	if fv.Metadata["name"] != "Test Corp" {
		t.Errorf("name metadata = %v, want Test Corp", fv.Metadata["name"])
	}
	if fv.Metadata["purpose"] != "buyer" {
		t.Errorf("purpose metadata = %v, want buyer", fv.Metadata["purpose"])
	}
}

func TestMAFeatureMapper_ToFeatureVector_NoFinancials(t *testing.T) {
	mapper := NewMAFeatureMapper()

	company := &domain.Company{
		ID:              "company-123",
		Name:            "Test Corp",
		Industry:        domain.IndustryFinance,
		Location:        "US",
		EmployeeCount:   100,
		Founded:         time.Now(),
		ListingStatus:   domain.ListingPrivate,
		MatchingPurpose: domain.PurposeSeller,
	}

	criteria := &domain.MAMatchingCriteria{
		CompanyID: "company-123",
		Purpose:   domain.PurposeSeller,
	}

	fv := mapper.ToFeatureVector(company, []*domain.Financials{}, criteria)

	// 基本プロパティは設定される
	if fv.ID != "company-123" {
		t.Errorf("ID = %v, want company-123", fv.ID)
	}
	if fv.Type != "ma_company" {
		t.Errorf("Type = %v, want ma_company", fv.Type)
	}

	// 財務情報がないため、数値特徴はない
	if len(fv.Numerical) != 0 {
		t.Errorf("Numerical features should be empty, got %d features", len(fv.Numerical))
	}

	// メタデータは設定される
	if fv.Metadata["name"] != "Test Corp" {
		t.Errorf("name metadata = %v, want Test Corp", fv.Metadata["name"])
	}
	if fv.Metadata["purpose"] != "seller" {
		t.Errorf("purpose metadata = %v, want seller", fv.Metadata["purpose"])
	}
}

func TestMAFeatureMapper_ToFeatureVector_SingleFinancial(t *testing.T) {
	mapper := NewMAFeatureMapper()

	company := &domain.Company{
		ID:              "company-123",
		Name:            "Test Corp",
		Industry:        domain.IndustryHealthcare,
		Location:        "GB",
		EmployeeCount:   1000,
		Founded:         time.Now(),
		ListingStatus:   domain.ListingPublic,
		MatchingPurpose: domain.PurposeBuyer,
	}

	financials := []*domain.Financials{
		{
			ID:               1,
			CompanyID:        "company-123",
			FiscalYear:       2024,
			Revenue:          5000000000,
			EBITDA:           1000000000,
			NetIncome:        750000000,
			TotalAssets:      25000000000,
			TotalLiabilities: 15000000000,
			Equity:           10000000000,
			ROE:              7.5,
			ROA:              3.0,
			DebtEquityRatio:  1.5,
			CurrentRatio:     2.0,
		},
	}

	criteria := &domain.MAMatchingCriteria{
		CompanyID: "company-123",
		Purpose:   domain.PurposeBuyer,
	}

	fv := mapper.ToFeatureVector(company, financials, criteria)

	// 数値特徴は設定される
	if len(fv.Numerical) != 7 {
		t.Errorf("Numerical features count = %d, want 7", len(fv.Numerical))
	}

	// 時系列特徴はない（1年分のデータのみ）
	if len(fv.TimeSeries) != 0 {
		t.Errorf("TimeSeries features should be empty with single year, got %d", len(fv.TimeSeries))
	}
}

func TestNormalizeRevenue(t *testing.T) {
	tests := []struct {
		name    string
		revenue int64
		want    func(float64) bool
	}{
		{
			name:    "小規模企業（1億円）",
			revenue: 100000000,
			want:    func(v float64) bool { return v >= 0 && v <= 1 },
		},
		{
			name:    "中規模企業（100億円）",
			revenue: 10000000000,
			want:    func(v float64) bool { return v >= 0 && v <= 1 },
		},
		{
			name:    "大規模企業（1兆円）",
			revenue: 1000000000000,
			want:    func(v float64) bool { return v >= 0 && v <= 1 },
		},
		{
			name:    "ゼロ売上",
			revenue: 0,
			want:    func(v float64) bool { return v == 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRevenue(tt.revenue)
			if !tt.want(got) {
				t.Errorf("normalizeRevenue(%d) = %v, validation failed", tt.revenue, got)
			}
		})
	}
}

func TestNormalizeMargin(t *testing.T) {
	tests := []struct {
		name    string
		ebitda  int64
		revenue int64
		want    float64
	}{
		{
			name:    "正常なマージン（20%） - マッピング後0.7",
			ebitda:  2000000000,
			revenue: 10000000000,
			want:    0.7, // (0.2 - (-0.5)) / (0.5 - (-0.5)) = 0.7
		},
		{
			name:    "高マージン（50%） - マッピング後1.0",
			ebitda:  5000000000,
			revenue: 10000000000,
			want:    1.0, // (0.5 - (-0.5)) / (0.5 - (-0.5)) = 1.0
		},
		{
			name:    "低マージン（5%） - マッピング後0.55",
			ebitda:  500000000,
			revenue: 10000000000,
			want:    0.55, // (0.05 - (-0.5)) / (0.5 - (-0.5)) = 0.55
		},
		{
			name:    "ゼロ売上 - デフォルト0.5",
			ebitda:  1000000000,
			revenue: 0,
			want:    0.5,
		},
		{
			name:    "負のEBITDA（-10%） - マッピング後0.4",
			ebitda:  -1000000000,
			revenue: 10000000000,
			want:    0.4, // (-0.1 - (-0.5)) / (0.5 - (-0.5)) = 0.4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMargin(tt.ebitda, tt.revenue)
			if got != tt.want {
				t.Errorf("normalizeMargin(%d, %d) = %v, want %v", tt.ebitda, tt.revenue, got, tt.want)
			}
		})
	}
}

func TestNormalizeROE(t *testing.T) {
	tests := []struct {
		name string
		roe  float64
		want func(float64) bool
	}{
		{
			name: "正常なROE（15%） - 0.15入力",
			roe:  0.15,
			want: func(v float64) bool { return v >= 0.4 && v <= 0.45 }, // (0.15 - (-0.5)) / 1.5 ≈ 0.433
		},
		{
			name: "高ROE（50%） - 0.5入力",
			roe:  0.5,
			want: func(v float64) bool { return v >= 0.65 && v <= 0.70 }, // (0.5 - (-0.5)) / 1.5 ≈ 0.667
		},
		{
			name: "負のROE（-50%） - -0.5入力でmin境界",
			roe:  -0.5,
			want: func(v float64) bool { return v == 0 },
		},
		{
			name: "ゼロROE - 0.0入力",
			roe:  0.0,
			want: func(v float64) bool { return v >= 0.33 && v <= 0.34 }, // (0 - (-0.5)) / 1.5 ≈ 0.333
		},
		{
			name: "最大ROE（100%） - 1.0入力でmax境界",
			roe:  1.0,
			want: func(v float64) bool { return v == 1.0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeROE(tt.roe)
			if !tt.want(got) {
				t.Errorf("normalizeROE(%v) = %v, validation failed", tt.roe, got)
			}
		})
	}
}

func TestNormalizeDebtEquity(t *testing.T) {
	tests := []struct {
		name string
		de   float64
		want func(float64) bool
	}{
		{
			name: "健全な負債比率（1.0）",
			de:   1.0,
			want: func(v float64) bool { return v >= 0 && v <= 1 },
		},
		{
			name: "高負債比率（3.0）",
			de:   3.0,
			want: func(v float64) bool { return v >= 0 && v <= 1 },
		},
		{
			name: "低負債比率（0.5）",
			de:   0.5,
			want: func(v float64) bool { return v >= 0 && v <= 1 },
		},
		{
			name: "負の比率",
			de:   -1.0,
			want: func(v float64) bool { return v == 0 },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeDebtEquity(tt.de)
			if !tt.want(got) {
				t.Errorf("normalizeDebtEquity(%v) = %v, validation failed", tt.de, got)
			}
		})
	}
}

func TestInferCompanyStage(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		company    *domain.Company
		financials *domain.Financials
		want       string
	}{
		{
			name: "スタートアップ（創業3年、低収益）",
			company: &domain.Company{
				Founded:       now.AddDate(-3, 0, 0),
				EmployeeCount: 50,
			},
			financials: &domain.Financials{
				Revenue: 100000000,
				ROE:     -5.0,
			},
			want: "startup",
		},
		{
			name: "成長期（創業7年、高ROE）",
			company: &domain.Company{
				Founded:       now.AddDate(-7, 0, 0),
				EmployeeCount: 300,
			},
			financials: &domain.Financials{
				Revenue: 5000000000,
				ROE:     20.0,
			},
			want: "growth",
		},
		{
			name: "成熟期（創業20年、従業員多数）",
			company: &domain.Company{
				Founded:       now.AddDate(-20, 0, 0),
				EmployeeCount: 5000,
			},
			financials: &domain.Financials{
				Revenue: 100000000000,
				ROE:     10.0,
			},
			want: "mature",
		},
		{
			name: "ターンアラウンド（負のROE）",
			company: &domain.Company{
				Founded:       now.AddDate(-10, 0, 0),
				EmployeeCount: 500,
			},
			financials: &domain.Financials{
				Revenue: 10000000000,
				ROE:     -8.0,
			},
			want: "turnaround",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferCompanyStage(tt.company, tt.financials)
			if got != tt.want {
				t.Errorf("inferCompanyStage() = %v, want %v", got, tt.want)
			}
		})
	}
}
