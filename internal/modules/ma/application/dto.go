package application

// MAMatchResult M&Aマッチング結果DTO
type MAMatchResult struct {
	CompanyID        string             `json:"company_id"`
	Score            float64            `json:"score"`
	Rank             int                `json:"rank"`
	Breakdown        map[string]float64 `json:"breakdown"`
	FinancialSummary *FinancialSummary  `json:"financial_summary"`
	SynergySummary   *SynergySummary    `json:"synergy_summary,omitempty"`
}

// FinancialSummary 財務サマリー
type FinancialSummary struct {
	Revenue         int64   `json:"revenue"`
	EBITDA          int64   `json:"ebitda"`
	EBITDAMargin    float64 `json:"ebitda_margin"`
	ROE             float64 `json:"roe"`
	ROA             float64 `json:"roa"`
	DebtEquityRatio float64 `json:"debt_equity_ratio"`
}

// SynergySummary シナジーサマリー
type SynergySummary struct {
	Type            string  `json:"type"`
	ExpectedSynergy float64 `json:"expected_synergy"`
	TechnologyFit   float64 `json:"technology_fit"`
	CustomerFit     float64 `json:"customer_fit"`
	OperationalFit  float64 `json:"operational_fit"`
}

// ValuationResult バリュエーション結果DTO
type ValuationResult struct {
	Method   string            `json:"method"`
	ValueMin float64           `json:"value_min"`
	ValueMax float64           `json:"value_max"`
	ValueMid float64           `json:"value_mid"`
	Multiple *IndustryMultiple `json:"multiple"`
	EBITDA   int64             `json:"ebitda"`
}

// IndustryMultiple 業界別マルチプル
type IndustryMultiple struct {
	Industry string  `json:"industry"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Median   float64 `json:"median"`
}

// InterestRequest 興味表明リクエスト
type InterestRequest struct {
	TargetCompanyID string `json:"target_company_id"`
}

// InterestResponse 興味表明レスポンス
type InterestResponse struct {
	Matched bool   `json:"matched"`
	MatchID string `json:"match_id,omitempty"`
}
