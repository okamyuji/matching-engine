package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// MAMatch M&Aマッチエンティティ
type MAMatch struct {
	bun.BaseModel `bun:"table:ma_matches"`

	ID             string             `bun:"id,pk"`
	CompanyIDA     string             `bun:"company_id_a,notnull"`
	CompanyIDB     string             `bun:"company_id_b,notnull"`
	Score          float64            `bun:"score,notnull"`
	Breakdown      map[string]float64 `bun:"breakdown,type:json"`
	SynergySummary *SynergySummary    `bun:"synergy_summary,type:json"`
	MatchedAt      time.Time          `bun:"matched_at,nullzero,default:current_timestamp"`
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
	bun.BaseModel `bun:"table:ma_interests"`

	ID            string    `bun:"id,pk"`
	FromCompanyID string    `bun:"from_company_id,notnull"`
	ToCompanyID   string    `bun:"to_company_id,notnull"`
	CreatedAt     time.Time `bun:"created_at,nullzero,default:current_timestamp"`
}
