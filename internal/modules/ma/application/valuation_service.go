package application

import (
	"context"
	"fmt"

	"github.com/yourorg/matching-engine/internal/modules/ma/domain"
	"github.com/yourorg/matching-engine/internal/modules/ma/infrastructure/repository"
)

// ValuationService バリュエーションサービス
type ValuationService struct {
	financialsRepo repository.FinancialsRepository
}

// NewValuationService 新しいValuationServiceを作成する
func NewValuationService(financialsRepo repository.FinancialsRepository) *ValuationService {
	return &ValuationService{
		financialsRepo: financialsRepo,
	}
}

// CalculateEVEBITDA EV/EBITDA法でバリュエーションを計算する
func (s *ValuationService) CalculateEVEBITDA(
	ctx context.Context,
	company *domain.Company,
) (*ValuationResult, error) {
	// 1. 最新財務情報取得
	latest, err := s.financialsRepo.FindLatest(ctx, company.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find latest financials: %w", err)
	}

	if latest.EBITDA <= 0 {
		return nil, fmt.Errorf("EBITDA must be positive for valuation")
	}

	// 2. 業界別マルチプル取得
	multiple := getIndustryMultiple(company.Industry)

	// 3. 企業価値（EV）レンジ計算
	ebitda := float64(latest.EBITDA)
	valueMin := ebitda * multiple.Min
	valueMax := ebitda * multiple.Max
	valueMid := ebitda * multiple.Median

	return &ValuationResult{
		Method:   "EV/EBITDA",
		ValueMin: valueMin,
		ValueMax: valueMax,
		ValueMid: valueMid,
		Multiple: multiple,
		EBITDA:   latest.EBITDA,
	}, nil
}

// getIndustryMultiple 業界別のEV/EBITDAマルチプルを取得する
func getIndustryMultiple(industry domain.Industry) *IndustryMultiple {
	// 業界別の典型的なEV/EBITDAマルチプル
	multiples := map[domain.Industry]*IndustryMultiple{
		domain.IndustryTechnology: {
			Industry: string(domain.IndustryTechnology),
			Min:      8.0,
			Max:      15.0,
			Median:   11.0,
		},
		domain.IndustryFinance: {
			Industry: string(domain.IndustryFinance),
			Min:      6.0,
			Max:      12.0,
			Median:   9.0,
		},
		domain.IndustryHealthcare: {
			Industry: string(domain.IndustryHealthcare),
			Min:      10.0,
			Max:      18.0,
			Median:   14.0,
		},
		domain.IndustryManufacturing: {
			Industry: string(domain.IndustryManufacturing),
			Min:      5.0,
			Max:      10.0,
			Median:   7.5,
		},
		domain.IndustryRetail: {
			Industry: string(domain.IndustryRetail),
			Min:      4.0,
			Max:      9.0,
			Median:   6.5,
		},
		domain.IndustryRealEstate: {
			Industry: string(domain.IndustryRealEstate),
			Min:      7.0,
			Max:      13.0,
			Median:   10.0,
		},
		domain.IndustryEnergy: {
			Industry: string(domain.IndustryEnergy),
			Min:      5.0,
			Max:      11.0,
			Median:   8.0,
		},
		domain.IndustryEducation: {
			Industry: string(domain.IndustryEducation),
			Min:      6.0,
			Max:      12.0,
			Median:   9.0,
		},
		domain.IndustryEntertainment: {
			Industry: string(domain.IndustryEntertainment),
			Min:      7.0,
			Max:      14.0,
			Median:   10.5,
		},
		domain.IndustryLogistics: {
			Industry: string(domain.IndustryLogistics),
			Min:      6.0,
			Max:      11.0,
			Median:   8.5,
		},
	}

	if multiple, ok := multiples[industry]; ok {
		return multiple
	}

	// デフォルトマルチプル
	return &IndustryMultiple{
		Industry: string(industry),
		Min:      6.0,
		Max:      12.0,
		Median:   9.0,
	}
}
