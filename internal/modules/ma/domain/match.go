package domain

import "time"

// MAMatch M&Aマッチエンティティ
type MAMatch struct {
	ID             string
	CompanyIDA     string
	CompanyIDB     string
	Score          float64
	Breakdown      map[string]float64
	SynergySummary *SynergySummary
	MatchedAt      time.Time
}

// SynergySummary シナジーサマリー
type SynergySummary struct {
	Type            SynergyType `json:"type"`
	ExpectedSynergy float64     `json:"expected_synergy"`
	TechnologyFit   float64     `json:"technology_fit"`
	CustomerFit     float64     `json:"customer_fit"`
	OperationalFit  float64     `json:"operational_fit"`
}

// Interest M&A興味表明エンティティ
type Interest struct {
	ID            string
	FromCompanyID string
	ToCompanyID   string
	CreatedAt     time.Time
}
