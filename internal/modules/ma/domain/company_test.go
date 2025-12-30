package domain

import (
	"testing"
	"time"
)

func TestCompany_Validation(t *testing.T) {
	founded := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

	c := &Company{
		ID:              "company-123",
		Name:            "Test Corp",
		Industry:        IndustryTechnology,
		Location:        "JP",
		EmployeeCount:   100,
		Founded:         founded,
		ListingStatus:   ListingPublic,
		MatchingPurpose: PurposeBuyer,
	}

	if c.ID != "company-123" {
		t.Errorf("ID = %v, want %v", c.ID, "company-123")
	}
	if c.Name != "Test Corp" {
		t.Errorf("Name = %v, want %v", c.Name, "Test Corp")
	}
	if c.Industry != IndustryTechnology {
		t.Errorf("Industry = %v, want %v", c.Industry, IndustryTechnology)
	}
	if c.Location != "JP" {
		t.Errorf("Location = %v, want %v", c.Location, "JP")
	}
	if c.EmployeeCount != 100 {
		t.Errorf("EmployeeCount = %v, want %v", c.EmployeeCount, 100)
	}
	if !c.Founded.Equal(founded) {
		t.Errorf("Founded = %v, want %v", c.Founded, founded)
	}
	if c.ListingStatus != ListingPublic {
		t.Errorf("ListingStatus = %v, want %v", c.ListingStatus, ListingPublic)
	}
	if c.MatchingPurpose != PurposeBuyer {
		t.Errorf("MatchingPurpose = %v, want %v", c.MatchingPurpose, PurposeBuyer)
	}
}

func TestCompany_Relations(t *testing.T) {
	c := &Company{}

	// Financials relation
	c.Financials = []*Financials{
		{
			ID:         1,
			CompanyID:  "company-123",
			FiscalYear: 2024,
			Revenue:    1000000,
			EBITDA:     200000,
		},
	}

	if len(c.Financials) != 1 {
		t.Errorf("len(Financials) = %v, want %v", len(c.Financials), 1)
	}
	if c.Financials[0].CompanyID != "company-123" {
		t.Errorf("Financials[0].CompanyID = %v, want %v", c.Financials[0].CompanyID, "company-123")
	}

	// Criteria relation
	c.Criteria = &MAMatchingCriteria{
		CompanyID: "company-123",
		Purpose:   PurposeBuyer,
	}

	if c.Criteria.CompanyID != "company-123" {
		t.Errorf("Criteria.CompanyID = %v, want %v", c.Criteria.CompanyID, "company-123")
	}

	// Technologies relation
	c.Technologies = []*CompanyTechnology{
		{
			CompanyID:  "company-123",
			Technology: "AI",
		},
	}

	if len(c.Technologies) != 1 {
		t.Errorf("len(Technologies) = %v, want %v", len(c.Technologies), 1)
	}

	// Markets relation
	c.Markets = []*CompanyMarket{
		{
			CompanyID: "company-123",
			Market:    "B2B",
		},
	}

	if len(c.Markets) != 1 {
		t.Errorf("len(Markets) = %v, want %v", len(c.Markets), 1)
	}
}

func TestIndustry_Constants(t *testing.T) {
	industries := []Industry{
		IndustryTechnology,
		IndustryFinance,
		IndustryHealthcare,
		IndustryManufacturing,
		IndustryRetail,
		IndustryRealEstate,
		IndustryEnergy,
		IndustryEducation,
		IndustryEntertainment,
		IndustryLogistics,
	}

	expected := []string{
		"technology",
		"finance",
		"healthcare",
		"manufacturing",
		"retail",
		"real_estate",
		"energy",
		"education",
		"entertainment",
		"logistics",
	}

	for i, industry := range industries {
		if string(industry) != expected[i] {
			t.Errorf("Industry constant %d = %v, want %v", i, string(industry), expected[i])
		}
	}
}

func TestListingStatus_Constants(t *testing.T) {
	if string(ListingPublic) != "public" {
		t.Errorf("ListingPublic = %v, want %v", string(ListingPublic), "public")
	}
	if string(ListingPrivate) != "private" {
		t.Errorf("ListingPrivate = %v, want %v", string(ListingPrivate), "private")
	}
}

func TestMatchingPurpose_Constants(t *testing.T) {
	if string(PurposeBuyer) != "buyer" {
		t.Errorf("PurposeBuyer = %v, want %v", string(PurposeBuyer), "buyer")
	}
	if string(PurposeSeller) != "seller" {
		t.Errorf("PurposeSeller = %v, want %v", string(PurposeSeller), "seller")
	}
}

func TestCompanyStage_Constants(t *testing.T) {
	stages := []CompanyStage{
		StageStartup,
		StageGrowth,
		StageMature,
		StageTurnaround,
	}

	expected := []string{
		"startup",
		"growth",
		"mature",
		"turnaround",
	}

	for i, stage := range stages {
		if string(stage) != expected[i] {
			t.Errorf("CompanyStage constant %d = %v, want %v", i, string(stage), expected[i])
		}
	}
}

func TestSynergyType_Constants(t *testing.T) {
	types := []SynergyType{
		SynergyHorizontal,
		SynergyVertical,
		SynergyDiversification,
		SynergyTechnology,
	}

	expected := []string{
		"horizontal_integration",
		"vertical_integration",
		"diversification",
		"technology_acquisition",
	}

	for i, synergyType := range types {
		if string(synergyType) != expected[i] {
			t.Errorf("SynergyType constant %d = %v, want %v", i, string(synergyType), expected[i])
		}
	}
}
