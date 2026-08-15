package application

import (
	"context"
	"errors"
	"testing"

	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
)

// MockFinancialsRepository FinancialsRepositoryのモック
type MockFinancialsRepository struct {
	FindLatestFunc func(ctx context.Context, companyID string) (*domain.Financials, error)
}

func (m *MockFinancialsRepository) FindLatest(ctx context.Context, companyID string) (*domain.Financials, error) {
	if m.FindLatestFunc != nil {
		return m.FindLatestFunc(ctx, companyID)
	}
	return nil, errors.New("FindLatestFunc not implemented")
}

func (m *MockFinancialsRepository) FindByCompanyID(ctx context.Context, companyID string, limit int) ([]*domain.Financials, error) {
	return nil, errors.New("not implemented")
}

func (m *MockFinancialsRepository) Save(ctx context.Context, financials *domain.Financials) error {
	return errors.New("not implemented")
}

func TestValuationService_CalculateEVEBITDA_Success(t *testing.T) {
	mockRepo := &MockFinancialsRepository{
		FindLatestFunc: func(ctx context.Context, companyID string) (*domain.Financials, error) {
			return &domain.Financials{
				ID:         1,
				CompanyID:  companyID,
				FiscalYear: 2024,
				Revenue:    10000000000,
				EBITDA:     2000000000, // 20億円
			}, nil
		},
	}

	service := NewValuationService(mockRepo)

	company := &domain.Company{
		ID:       "company-123",
		Industry: domain.IndustryTechnology,
	}

	result, err := service.CalculateEVEBITDA(context.Background(), company)

	if err != nil {
		t.Fatalf("CalculateEVEBITDA() error = %v", err)
	}

	if result == nil {
		t.Fatal("result should not be nil")
	}

	if result.Method != "EV/EBITDA" {
		t.Errorf("Method = %v, want EV/EBITDA", result.Method)
	}

	if result.EBITDA != 2000000000 {
		t.Errorf("EBITDA = %v, want 2000000000", result.EBITDA)
	}

	// Technology業界の場合、マルチプルが10-15x程度
	if result.ValueMin <= 0 {
		t.Errorf("ValueMin = %v, should be positive", result.ValueMin)
	}
	if result.ValueMax <= result.ValueMin {
		t.Errorf("ValueMax = %v, should be greater than ValueMin = %v", result.ValueMax, result.ValueMin)
	}
	if result.ValueMid < result.ValueMin || result.ValueMid > result.ValueMax {
		t.Errorf("ValueMid = %v, should be between ValueMin and ValueMax", result.ValueMid)
	}

	// マルチプルの確認
	if result.Multiple.Min <= 0 {
		t.Errorf("Multiple.Min = %v, should be positive", result.Multiple.Min)
	}
	if result.Multiple.Max <= result.Multiple.Min {
		t.Errorf("Multiple.Max = %v, should be greater than Multiple.Min", result.Multiple.Max)
	}
	if result.Multiple.Median < result.Multiple.Min || result.Multiple.Median > result.Multiple.Max {
		t.Errorf("Multiple.Median = %v, should be between Min and Max", result.Multiple.Median)
	}
}

func TestValuationService_CalculateEVEBITDA_ZeroEBITDA(t *testing.T) {
	mockRepo := &MockFinancialsRepository{
		FindLatestFunc: func(ctx context.Context, companyID string) (*domain.Financials, error) {
			return &domain.Financials{
				ID:         1,
				CompanyID:  companyID,
				FiscalYear: 2024,
				Revenue:    10000000000,
				EBITDA:     0, // ゼロEBITDA
			}, nil
		},
	}

	service := NewValuationService(mockRepo)

	company := &domain.Company{
		ID:       "company-123",
		Industry: domain.IndustryTechnology,
	}

	_, err := service.CalculateEVEBITDA(context.Background(), company)

	if err == nil {
		t.Error("CalculateEVEBITDA() should return error for zero EBITDA")
	}
}

func TestValuationService_CalculateEVEBITDA_NegativeEBITDA(t *testing.T) {
	mockRepo := &MockFinancialsRepository{
		FindLatestFunc: func(ctx context.Context, companyID string) (*domain.Financials, error) {
			return &domain.Financials{
				ID:         1,
				CompanyID:  companyID,
				FiscalYear: 2024,
				Revenue:    10000000000,
				EBITDA:     -1000000000, // 負のEBITDA
			}, nil
		},
	}

	service := NewValuationService(mockRepo)

	company := &domain.Company{
		ID:       "company-123",
		Industry: domain.IndustryTechnology,
	}

	_, err := service.CalculateEVEBITDA(context.Background(), company)

	if err == nil {
		t.Error("CalculateEVEBITDA() should return error for negative EBITDA")
	}
}

func TestValuationService_CalculateEVEBITDA_FinancialsNotFound(t *testing.T) {
	mockRepo := &MockFinancialsRepository{
		FindLatestFunc: func(ctx context.Context, companyID string) (*domain.Financials, error) {
			return nil, errors.New("financials not found")
		},
	}

	service := NewValuationService(mockRepo)

	company := &domain.Company{
		ID:       "company-123",
		Industry: domain.IndustryTechnology,
	}

	_, err := service.CalculateEVEBITDA(context.Background(), company)

	if err == nil {
		t.Error("CalculateEVEBITDA() should return error when financials not found")
	}
}

func TestValuationService_CalculateEVEBITDA_AllIndustries(t *testing.T) {
	industries := []domain.Industry{
		domain.IndustryTechnology,
		domain.IndustryFinance,
		domain.IndustryHealthcare,
		domain.IndustryManufacturing,
		domain.IndustryRetail,
		domain.IndustryRealEstate,
		domain.IndustryEnergy,
		domain.IndustryEducation,
		domain.IndustryEntertainment,
		domain.IndustryLogistics,
	}

	mockRepo := &MockFinancialsRepository{
		FindLatestFunc: func(ctx context.Context, companyID string) (*domain.Financials, error) {
			return &domain.Financials{
				ID:         1,
				CompanyID:  companyID,
				FiscalYear: 2024,
				Revenue:    10000000000,
				EBITDA:     2000000000,
			}, nil
		},
	}

	service := NewValuationService(mockRepo)

	for _, industry := range industries {
		t.Run(string(industry), func(t *testing.T) {
			company := &domain.Company{
				ID:       "company-123",
				Industry: industry,
			}

			result, err := service.CalculateEVEBITDA(context.Background(), company)

			if err != nil {
				t.Fatalf("CalculateEVEBITDA() error = %v for industry %v", err, industry)
			}

			if result == nil {
				t.Fatal("result should not be nil")
			}

			// 各業界でマルチプルが適切に設定されているか確認
			if result.Multiple.Min <= 0 || result.Multiple.Max <= 0 || result.Multiple.Median <= 0 {
				t.Errorf("Multiple values should be positive for industry %v: Min=%v, Max=%v, Median=%v",
					industry, result.Multiple.Min, result.Multiple.Max, result.Multiple.Median)
			}

			// バリュエーションが計算されているか確認
			if result.ValueMin <= 0 || result.ValueMax <= 0 || result.ValueMid <= 0 {
				t.Errorf("Valuation values should be positive for industry %v: Min=%v, Max=%v, Mid=%v",
					industry, result.ValueMin, result.ValueMax, result.ValueMid)
			}
		})
	}
}

func TestGetIndustryMultiple(t *testing.T) {
	tests := []struct {
		industry    domain.Industry
		checkMedian bool
		minMedian   float64
		maxMedian   float64
	}{
		{
			industry:    domain.IndustryTechnology,
			checkMedian: true,
			minMedian:   10.0,
			maxMedian:   15.0,
		},
		{
			industry:    domain.IndustryFinance,
			checkMedian: true,
			minMedian:   8.0,
			maxMedian:   12.0,
		},
		{
			industry:    domain.IndustryHealthcare,
			checkMedian: true,
			minMedian:   12.0,
			maxMedian:   18.0,
		},
		{
			industry:    domain.IndustryManufacturing,
			checkMedian: true,
			minMedian:   6.0,
			maxMedian:   10.0,
		},
		{
			industry:    domain.IndustryRetail,
			checkMedian: true,
			minMedian:   5.0,
			maxMedian:   9.0,
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.industry), func(t *testing.T) {
			multiple := getIndustryMultiple(tt.industry)

			if multiple.Min <= 0 {
				t.Errorf("Min = %v, should be positive", multiple.Min)
			}
			if multiple.Max <= multiple.Min {
				t.Errorf("Max = %v, should be greater than Min = %v", multiple.Max, multiple.Min)
			}
			if multiple.Median < multiple.Min || multiple.Median > multiple.Max {
				t.Errorf("Median = %v, should be between Min = %v and Max = %v",
					multiple.Median, multiple.Min, multiple.Max)
			}

			if tt.checkMedian {
				if multiple.Median < tt.minMedian || multiple.Median > tt.maxMedian {
					t.Errorf("Median = %v, expected between %v and %v",
						multiple.Median, tt.minMedian, tt.maxMedian)
				}
			}
		})
	}
}
