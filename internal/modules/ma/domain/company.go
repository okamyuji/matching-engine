package domain

import (
	"time"

	"github.com/uptrace/bun"
)

// Company 企業エンティティ
type Company struct {
	bun.BaseModel `bun:"table:ma_companies"`

	ID              string          `bun:"id,pk"`
	Name            string          `bun:"name,notnull"`
	Industry        Industry        `bun:"industry,notnull"`
	Location        string          `bun:"location,notnull"`
	EmployeeCount   int             `bun:"employee_count,notnull"`
	Founded         time.Time       `bun:"founded,notnull"`
	ListingStatus   ListingStatus   `bun:"listing_status,notnull,default:'private'"`
	MatchingPurpose MatchingPurpose `bun:"matching_purpose,notnull"`
	CreatedAt       time.Time       `bun:"created_at,nullzero,default:current_timestamp"`
	UpdatedAt       time.Time       `bun:"updated_at,nullzero,default:current_timestamp"`

	// リレーション
	Financials   []*Financials        `bun:"rel:has-many,join:id=company_id"`
	Criteria     *MAMatchingCriteria  `bun:"rel:has-one,join:id=company_id"`
	Technologies []*CompanyTechnology `bun:"rel:has-many,join:id=company_id"`
	Markets      []*CompanyMarket     `bun:"rel:has-many,join:id=company_id"`
}

// CompanyTechnology 企業が保有する技術
type CompanyTechnology struct {
	bun.BaseModel `bun:"table:ma_company_technologies"`

	ID         int64  `bun:"id,pk,autoincrement"`
	CompanyID  string `bun:"company_id,notnull"`
	Technology string `bun:"technology,notnull"`
}

// CompanyMarket 企業が展開する市場
type CompanyMarket struct {
	bun.BaseModel `bun:"table:ma_company_markets"`

	ID        int64  `bun:"id,pk,autoincrement"`
	CompanyID string `bun:"company_id,notnull"`
	Market    string `bun:"market,notnull"`
}
