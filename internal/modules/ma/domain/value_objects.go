package domain

// Industry 業界分類
type Industry string

const (
	IndustryTechnology    Industry = "technology"
	IndustryFinance       Industry = "finance"
	IndustryHealthcare    Industry = "healthcare"
	IndustryManufacturing Industry = "manufacturing"
	IndustryRetail        Industry = "retail"
	IndustryRealEstate    Industry = "real_estate"
	IndustryEnergy        Industry = "energy"
	IndustryEducation     Industry = "education"
	IndustryEntertainment Industry = "entertainment"
	IndustryLogistics     Industry = "logistics"
)

// ListingStatus 上場状態
type ListingStatus string

const (
	ListingPublic  ListingStatus = "public"
	ListingPrivate ListingStatus = "private"
)

// MatchingPurpose マッチング目的
type MatchingPurpose string

const (
	PurposeBuyer  MatchingPurpose = "buyer"
	PurposeSeller MatchingPurpose = "seller"
)

// CompanyStage 企業ステージ
type CompanyStage string

const (
	StageStartup    CompanyStage = "startup"
	StageGrowth     CompanyStage = "growth"
	StageMature     CompanyStage = "mature"
	StageTurnaround CompanyStage = "turnaround"
)

// SynergyType シナジータイプ
type SynergyType string

const (
	SynergyHorizontal      SynergyType = "horizontal_integration"
	SynergyVertical        SynergyType = "vertical_integration"
	SynergyDiversification SynergyType = "diversification"
	SynergyTechnology      SynergyType = "technology_acquisition"
)
