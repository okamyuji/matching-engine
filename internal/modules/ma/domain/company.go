package domain

import "time"

// Company 企業エンティティ
type Company struct {
	ID              string
	Name            string
	Industry        Industry
	Location        string
	EmployeeCount   int
	Founded         time.Time
	ListingStatus   ListingStatus
	MatchingPurpose MatchingPurpose
	CreatedAt       time.Time
	UpdatedAt       time.Time

	// リレーション
	Financials   []*Financials
	Criteria     *MAMatchingCriteria
	Technologies []*CompanyTechnology
	Markets      []*CompanyMarket
}

// CompanyTechnology 企業が保有する技術
type CompanyTechnology struct {
	ID         int64
	CompanyID  string
	Technology string
}

// CompanyMarket 企業が展開する市場
type CompanyMarket struct {
	ID        int64
	CompanyID string
	Market    string
}
