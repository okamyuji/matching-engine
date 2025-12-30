package domain

import (
	"testing"
	"time"
)

func TestMAMatch_Validation(t *testing.T) {
	match := &MAMatch{
		ID:         "match-123",
		CompanyIDA: "company-a",
		CompanyIDB: "company-b",
		Score:      0.85,
		Breakdown: map[string]float64{
			"financial_health":     0.90,
			"industry_synergy":     0.80,
			"technology_overlap":   0.85,
			"growth_trajectory":    0.88,
			"profitability_trend":  0.82,
			"customer_segment_fit": 0.75,
			"market_overlap":       0.70,
		},
	}

	if match.ID != "match-123" {
		t.Errorf("ID = %v, want %v", match.ID, "match-123")
	}
	if match.CompanyIDA != "company-a" {
		t.Errorf("CompanyIDA = %v, want %v", match.CompanyIDA, "company-a")
	}
	if match.CompanyIDB != "company-b" {
		t.Errorf("CompanyIDB = %v, want %v", match.CompanyIDB, "company-b")
	}
	if match.Score != 0.85 {
		t.Errorf("Score = %v, want %v", match.Score, 0.85)
	}
	if len(match.Breakdown) != 7 {
		t.Errorf("len(Breakdown) = %v, want %v", len(match.Breakdown), 7)
	}
}

func TestMAMatch_Breakdown(t *testing.T) {
	match := &MAMatch{
		Breakdown: map[string]float64{
			"financial_health":   0.90,
			"industry_synergy":   0.75,
			"technology_overlap": 0.85,
		},
	}

	// Breakdownの各要素を確認
	if match.Breakdown["financial_health"] != 0.90 {
		t.Errorf("Breakdown[financial_health] = %v, want %v", match.Breakdown["financial_health"], 0.90)
	}
	if match.Breakdown["industry_synergy"] != 0.75 {
		t.Errorf("Breakdown[industry_synergy] = %v, want %v", match.Breakdown["industry_synergy"], 0.75)
	}
	if match.Breakdown["technology_overlap"] != 0.85 {
		t.Errorf("Breakdown[technology_overlap] = %v, want %v", match.Breakdown["technology_overlap"], 0.85)
	}
}

func TestMAMatch_EmptyBreakdown(t *testing.T) {
	match := &MAMatch{
		Breakdown: make(map[string]float64),
	}

	if len(match.Breakdown) != 0 {
		t.Errorf("len(Breakdown) = %v, want %v", len(match.Breakdown), 0)
	}
}

func TestInterest_Validation(t *testing.T) {
	now := time.Now()

	interest := &Interest{
		ID:            "interest-123",
		FromCompanyID: "company-a",
		ToCompanyID:   "company-b",
		CreatedAt:     now,
	}

	if interest.ID != "interest-123" {
		t.Errorf("ID = %v, want %v", interest.ID, "interest-123")
	}
	if interest.FromCompanyID != "company-a" {
		t.Errorf("FromCompanyID = %v, want %v", interest.FromCompanyID, "company-a")
	}
	if interest.ToCompanyID != "company-b" {
		t.Errorf("ToCompanyID = %v, want %v", interest.ToCompanyID, "company-b")
	}
	if !interest.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", interest.CreatedAt, now)
	}
}

func TestInterest_MultipleInterests(t *testing.T) {
	// 複数の興味表明を作成
	interests := []*Interest{
		{
			ID:            "interest-1",
			FromCompanyID: "company-a",
			ToCompanyID:   "company-b",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest-2",
			FromCompanyID: "company-b",
			ToCompanyID:   "company-a",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest-3",
			FromCompanyID: "company-c",
			ToCompanyID:   "company-a",
			CreatedAt:     time.Now(),
		},
	}

	if len(interests) != 3 {
		t.Errorf("len(interests) = %v, want %v", len(interests), 3)
	}

	// 各興味表明が異なるIDを持つことを確認
	ids := make(map[string]bool)
	for _, interest := range interests {
		if ids[interest.ID] {
			t.Errorf("Duplicate ID found: %v", interest.ID)
		}
		ids[interest.ID] = true
	}
}

func TestCompanyTechnology_Validation(t *testing.T) {
	tech := &CompanyTechnology{
		CompanyID:  "company-123",
		Technology: "AI",
	}

	if tech.CompanyID != "company-123" {
		t.Errorf("CompanyID = %v, want %v", tech.CompanyID, "company-123")
	}
	if tech.Technology != "AI" {
		t.Errorf("Technology = %v, want %v", tech.Technology, "AI")
	}
}

func TestCompanyMarket_Validation(t *testing.T) {
	market := &CompanyMarket{
		CompanyID: "company-123",
		Market:    "B2B",
	}

	if market.CompanyID != "company-123" {
		t.Errorf("CompanyID = %v, want %v", market.CompanyID, "company-123")
	}
	if market.Market != "B2B" {
		t.Errorf("Market = %v, want %v", market.Market, "B2B")
	}
}

func TestMAMatch_ScoreBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		score float64
	}{
		{"最小スコア", 0.0},
		{"最大スコア", 1.0},
		{"中間スコア", 0.5},
		{"高スコア", 0.95},
		{"低スコア", 0.15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match := &MAMatch{
				Score: tt.score,
			}

			if match.Score != tt.score {
				t.Errorf("Score = %v, want %v", match.Score, tt.score)
			}
		})
	}
}

func TestSynergySummary(t *testing.T) {
	summary := &SynergySummary{
		Type:            SynergyHorizontal,
		ExpectedSynergy: 0.85,
		TechnologyFit:   0.90,
		CustomerFit:     0.80,
		OperationalFit:  0.85,
	}

	if summary.Type != SynergyHorizontal {
		t.Errorf("Type = %v, want %v", summary.Type, SynergyHorizontal)
	}
	if summary.ExpectedSynergy != 0.85 {
		t.Errorf("ExpectedSynergy = %v, want %v", summary.ExpectedSynergy, 0.85)
	}
	if summary.TechnologyFit != 0.90 {
		t.Errorf("TechnologyFit = %v, want %v", summary.TechnologyFit, 0.90)
	}
	if summary.CustomerFit != 0.80 {
		t.Errorf("CustomerFit = %v, want %v", summary.CustomerFit, 0.80)
	}
	if summary.OperationalFit != 0.85 {
		t.Errorf("OperationalFit = %v, want %v", summary.OperationalFit, 0.85)
	}
}

func TestMAMatch_WithSynergySummary(t *testing.T) {
	match := &MAMatch{
		SynergySummary: &SynergySummary{
			Type:            SynergyVertical,
			ExpectedSynergy: 0.80,
			TechnologyFit:   0.85,
			CustomerFit:     0.75,
			OperationalFit:  0.80,
		},
	}

	if match.SynergySummary == nil {
		t.Error("SynergySummary should not be nil")
		return
	}

	if match.SynergySummary.Type != SynergyVertical {
		t.Errorf("SynergySummary.Type = %v, want %v", match.SynergySummary.Type, SynergyVertical)
	}
	if match.SynergySummary.ExpectedSynergy != 0.80 {
		t.Errorf("SynergySummary.ExpectedSynergy = %v, want %v", match.SynergySummary.ExpectedSynergy, 0.80)
	}
}
