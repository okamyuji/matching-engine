package application

import (
	"testing"

	"github.com/yourorg/matching-engine/internal/core/matching"
)

func TestSynergyCalculator_Calculate(t *testing.T) {
	calc := NewSynergyCalculator()

	source := matching.NewFeatureVector("source-123", "ma_company")
	source.SetCategorical("industry", "technology", 1.0)
	source.SetSparse("technology", "AI", 1.0)
	source.SetSparse("technology", "Cloud", 1.0)
	source.SetSparse("market", "B2B", 1.0)

	candidate := matching.NewFeatureVector("candidate-456", "ma_company")
	candidate.SetCategorical("industry", "technology", 1.0)
	candidate.SetSparse("technology", "AI", 1.0)
	candidate.SetSparse("technology", "Blockchain", 1.0)
	candidate.SetSparse("market", "B2C", 1.0)

	summary := calc.Calculate(source, candidate)

	if summary == nil {
		t.Fatal("SynergySummary should not be nil")
	}

	// シナジータイプの確認（同業界+技術重複 → 水平統合）
	if summary.Type == "" {
		t.Error("SynergyType should not be empty")
	}

	// 期待シナジー値の確認
	if summary.ExpectedSynergy < 0 || summary.ExpectedSynergy > 1 {
		t.Errorf("ExpectedSynergy = %v, should be between 0 and 1", summary.ExpectedSynergy)
	}

	// フィット値の確認
	if summary.TechnologyFit < 0 || summary.TechnologyFit > 1 {
		t.Errorf("TechnologyFit = %v, should be between 0 and 1", summary.TechnologyFit)
	}
	if summary.CustomerFit < 0 || summary.CustomerFit > 1 {
		t.Errorf("CustomerFit = %v, should be between 0 and 1", summary.CustomerFit)
	}
	if summary.OperationalFit < 0 || summary.OperationalFit > 1 {
		t.Errorf("OperationalFit = %v, should be between 0 and 1", summary.OperationalFit)
	}
}

func TestSynergyCalculator_DetermineSynergyType_Horizontal(t *testing.T) {
	calc := NewSynergyCalculator()

	// 同業界 + 高い技術重複 → 水平統合
	source := matching.NewFeatureVector("source-123", "ma_company")
	source.SetCategorical("industry", "technology", 1.0)
	source.SetSparse("technology", "AI", 1.0)
	source.SetSparse("technology", "Cloud", 1.0)

	candidate := matching.NewFeatureVector("candidate-456", "ma_company")
	candidate.SetCategorical("industry", "technology", 1.0)
	candidate.SetSparse("technology", "AI", 1.0)
	candidate.SetSparse("technology", "Cloud", 1.0)

	synergyType := calc.determineSynergyType(source, candidate)

	if synergyType != "horizontal_integration" {
		t.Errorf("SynergyType = %v, want horizontal_integration", synergyType)
	}
}

func TestSynergyCalculator_DetermineSynergyType_Vertical(t *testing.T) {
	calc := NewSynergyCalculator()

	// 異業界 + 高い技術重複 → 垂直統合
	source := matching.NewFeatureVector("source-123", "ma_company")
	source.SetCategorical("industry", "technology", 1.0)
	source.SetSparse("technology", "AI", 1.0)
	source.SetSparse("technology", "Cloud", 1.0)

	candidate := matching.NewFeatureVector("candidate-456", "ma_company")
	candidate.SetCategorical("industry", "finance", 1.0)
	candidate.SetSparse("technology", "AI", 1.0)
	candidate.SetSparse("technology", "Cloud", 1.0)

	synergyType := calc.determineSynergyType(source, candidate)

	if synergyType != "vertical_integration" {
		t.Errorf("SynergyType = %v, want vertical_integration", synergyType)
	}
}

func TestSynergyCalculator_DetermineSynergyType_Technology(t *testing.T) {
	calc := NewSynergyCalculator()

	// 同業界 + 低い技術重複 → 技術獲得
	source := matching.NewFeatureVector("source-123", "ma_company")
	source.SetCategorical("industry", "technology", 1.0)
	source.SetSparse("technology", "AI", 1.0)

	candidate := matching.NewFeatureVector("candidate-456", "ma_company")
	candidate.SetCategorical("industry", "technology", 1.0)
	candidate.SetSparse("technology", "Blockchain", 1.0)

	synergyType := calc.determineSynergyType(source, candidate)

	if synergyType != "technology_acquisition" {
		t.Errorf("SynergyType = %v, want technology_acquisition", synergyType)
	}
}

func TestSynergyCalculator_DetermineSynergyType_Diversification(t *testing.T) {
	calc := NewSynergyCalculator()

	// 異業界 + 低い技術重複 → 多角化
	source := matching.NewFeatureVector("source-123", "ma_company")
	source.SetCategorical("industry", "technology", 1.0)
	source.SetSparse("technology", "AI", 1.0)

	candidate := matching.NewFeatureVector("candidate-456", "ma_company")
	candidate.SetCategorical("industry", "retail", 1.0)
	candidate.SetSparse("technology", "Logistics", 1.0)

	synergyType := calc.determineSynergyType(source, candidate)

	if synergyType != "diversification" {
		t.Errorf("SynergyType = %v, want diversification", synergyType)
	}
}

func TestSynergyCalculator_CalculateTechnologyFit(t *testing.T) {
	calc := NewSynergyCalculator()

	tests := []struct {
		name          string
		sourceTech    map[string]float64
		candidateTech map[string]float64
		expectMinFit  float64
		expectMaxFit  float64
	}{
		{
			name:          "完全一致（補完性低い）",
			sourceTech:    map[string]float64{"AI": 1.0, "Cloud": 1.0},
			candidateTech: map[string]float64{"AI": 1.0, "Cloud": 1.0},
			expectMinFit:  0.0,
			expectMaxFit:  0.2, // 重複が多いので補完性は低い
		},
		{
			name:          "部分一致（適度な補完性）",
			sourceTech:    map[string]float64{"AI": 1.0, "Cloud": 1.0},
			candidateTech: map[string]float64{"AI": 1.0, "Blockchain": 1.0},
			expectMinFit:  0.3,
			expectMaxFit:  0.7,
		},
		{
			name:          "一致なし（補完性高い）",
			sourceTech:    map[string]float64{"AI": 1.0},
			candidateTech: map[string]float64{"Blockchain": 1.0},
			expectMinFit:  0.6,
			expectMaxFit:  0.8, // 重複がないので補完性は高い
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := matching.NewFeatureVector("source", "ma_company")
			for tech, val := range tt.sourceTech {
				source.SetSparse("technology", tech, val)
			}

			candidate := matching.NewFeatureVector("candidate", "ma_company")
			for tech, val := range tt.candidateTech {
				candidate.SetSparse("technology", tech, val)
			}

			fit := calc.calculateTechnologyFit(source, candidate)

			if fit < tt.expectMinFit || fit > tt.expectMaxFit {
				t.Errorf("TechnologyFit = %v, want between %v and %v", fit, tt.expectMinFit, tt.expectMaxFit)
			}
		})
	}
}

func TestSynergyCalculator_CalculateCustomerFit(t *testing.T) {
	calc := NewSynergyCalculator()

	source := matching.NewFeatureVector("source", "ma_company")
	source.SetSparse("market", "B2B", 1.0)
	source.SetSparse("market", "Enterprise", 1.0)

	candidate := matching.NewFeatureVector("candidate", "ma_company")
	candidate.SetSparse("market", "B2B", 1.0)
	candidate.SetSparse("market", "SMB", 1.0)

	fit := calc.calculateCustomerFit(source, candidate)

	if fit < 0 || fit > 1 {
		t.Errorf("CustomerFit = %v, should be between 0 and 1", fit)
	}
}

func TestSynergyCalculator_CalculateOperationalFit(t *testing.T) {
	calc := NewSynergyCalculator()

	// 同じ業界
	source := matching.NewFeatureVector("source", "ma_company")
	source.SetCategorical("industry", "technology", 1.0)

	candidate := matching.NewFeatureVector("candidate", "ma_company")
	candidate.SetCategorical("industry", "technology", 1.0)

	fit := calc.calculateOperationalFit(source, candidate)

	if fit < 0.7 || fit > 1.0 {
		t.Errorf("OperationalFit for same industry = %v, should be between 0.7 and 1.0", fit)
	}
}

func TestSynergyCalculator_CalculateOperationalFit_DifferentIndustry(t *testing.T) {
	calc := NewSynergyCalculator()

	// 異なる業界
	source := matching.NewFeatureVector("source", "ma_company")
	source.SetCategorical("industry", "technology", 1.0)

	candidate := matching.NewFeatureVector("candidate", "ma_company")
	candidate.SetCategorical("industry", "finance", 1.0)

	fit := calc.calculateOperationalFit(source, candidate)

	if fit < 0.0 || fit > 0.5 {
		t.Errorf("OperationalFit for different industry = %v, should be between 0.0 and 0.5", fit)
	}
}

func TestSynergyCalculator_CalculateExpectedSynergy(t *testing.T) {
	calc := NewSynergyCalculator()

	tests := []struct {
		name           string
		synergyType    string
		techFit        float64
		customerFit    float64
		operationalFit float64
	}{
		{
			name:           "水平統合",
			synergyType:    "horizontal_integration",
			techFit:        0.8,
			customerFit:    0.7,
			operationalFit: 0.9,
		},
		{
			name:           "垂直統合",
			synergyType:    "vertical_integration",
			techFit:        0.9,
			customerFit:    0.6,
			operationalFit: 0.7,
		},
		{
			name:           "多角化",
			synergyType:    "diversification",
			techFit:        0.3,
			customerFit:    0.4,
			operationalFit: 0.3,
		},
		{
			name:           "技術獲得",
			synergyType:    "technology_acquisition",
			techFit:        0.5,
			customerFit:    0.8,
			operationalFit: 0.7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synergy := calc.calculateExpectedSynergy(
				tt.synergyType,
				tt.techFit,
				tt.customerFit,
				tt.operationalFit,
			)

			if synergy < 0 || synergy > 1 {
				t.Errorf("ExpectedSynergy = %v, should be between 0 and 1", synergy)
			}
		})
	}
}
