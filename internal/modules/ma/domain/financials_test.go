package domain

import (
	"testing"
)

func TestFinancials_EBITDAMargin(t *testing.T) {
	tests := []struct {
		name    string
		revenue int64
		ebitda  int64
		want    float64
	}{
		{
			name:    "正常なEBITDAマージン計算",
			revenue: 1000000,
			ebitda:  200000,
			want:    0.2,
		},
		{
			name:    "売上ゼロの場合はゼロを返す",
			revenue: 0,
			ebitda:  100000,
			want:    0.0,
		},
		{
			name:    "負のEBITDA",
			revenue: 1000000,
			ebitda:  -50000,
			want:    -0.05,
		},
		{
			name:    "高利益率",
			revenue: 500000,
			ebitda:  250000,
			want:    0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Financials{
				Revenue: tt.revenue,
				EBITDA:  tt.ebitda,
			}

			got := f.EBITDAMargin()
			if got != tt.want {
				t.Errorf("EBITDAMargin() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFinancials_IsHealthy(t *testing.T) {
	tests := []struct {
		name        string
		financials  *Financials
		want        bool
		description string
	}{
		{
			name: "健全な財務状態",
			financials: &Financials{
				ROE:             15.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 1.5,
				CurrentRatio:    2.0,
			},
			want:        true,
			description: "全ての指標が健全",
		},
		{
			name: "ROEが負",
			financials: &Financials{
				ROE:             -5.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 1.5,
				CurrentRatio:    2.0,
			},
			want:        false,
			description: "ROEが0以下",
		},
		{
			name: "ROAが負",
			financials: &Financials{
				ROE:             15.0,
				ROA:             -2.0,
				NetIncome:       100000,
				DebtEquityRatio: 1.5,
				CurrentRatio:    2.0,
			},
			want:        false,
			description: "ROAが0以下",
		},
		{
			name: "純利益が負",
			financials: &Financials{
				ROE:             15.0,
				ROA:             8.0,
				NetIncome:       -50000,
				DebtEquityRatio: 1.5,
				CurrentRatio:    2.0,
			},
			want:        false,
			description: "純利益が0以下",
		},
		{
			name: "負債比率が高すぎる",
			financials: &Financials{
				ROE:             15.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 2.5,
				CurrentRatio:    2.0,
			},
			want:        false,
			description: "デット・エクイティ・レシオが2.0以上",
		},
		{
			name: "流動比率が低すぎる",
			financials: &Financials{
				ROE:             15.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 1.5,
				CurrentRatio:    0.8,
			},
			want:        false,
			description: "流動比率が1.0以下",
		},
		{
			name: "複数の指標が問題",
			financials: &Financials{
				ROE:             -10.0,
				ROA:             -5.0,
				NetIncome:       -100000,
				DebtEquityRatio: 3.0,
				CurrentRatio:    0.5,
			},
			want:        false,
			description: "全ての指標が不健全",
		},
		{
			name: "境界値 - ROEゼロ",
			financials: &Financials{
				ROE:             0.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 1.5,
				CurrentRatio:    2.0,
			},
			want:        false,
			description: "ROEがちょうど0",
		},
		{
			name: "境界値 - デット・エクイティ・レシオが2.0",
			financials: &Financials{
				ROE:             15.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 2.0,
				CurrentRatio:    2.0,
			},
			want:        false,
			description: "デット・エクイティ・レシオがちょうど2.0",
		},
		{
			name: "境界値 - 流動比率が1.0",
			financials: &Financials{
				ROE:             15.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 1.5,
				CurrentRatio:    1.0,
			},
			want:        false,
			description: "流動比率がちょうど1.0",
		},
		{
			name: "境界値 - 流動比率が1.1（健全）",
			financials: &Financials{
				ROE:             15.0,
				ROA:             8.0,
				NetIncome:       100000,
				DebtEquityRatio: 1.9,
				CurrentRatio:    1.1,
			},
			want:        true,
			description: "全ての指標がわずかに基準を満たす",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.financials.IsHealthy()
			if got != tt.want {
				t.Errorf("IsHealthy() = %v, want %v (case: %s)", got, tt.want, tt.description)
			}
		})
	}
}

func TestFinancials_Validation(t *testing.T) {
	// フィールドが正しく設定できることを確認
	f := &Financials{
		CompanyID:  "company-123",
		FiscalYear: 2024,
		Revenue:    10000000,
	}

	if f.CompanyID != "company-123" {
		t.Errorf("CompanyID = %v, want %v", f.CompanyID, "company-123")
	}
	if f.FiscalYear != 2024 {
		t.Errorf("FiscalYear = %v, want %v", f.FiscalYear, 2024)
	}
	if f.Revenue != 10000000 {
		t.Errorf("Revenue = %v, want %v", f.Revenue, 10000000)
	}
}
