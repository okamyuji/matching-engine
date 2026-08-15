package mapper

import (
	"math"
	"time"

	"github.com/okamyuji/matching-engine/internal/core/matching"
	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
)

// MAFeatureMapper M&Aドメインから特徴ベクトルへの変換
type MAFeatureMapper struct{}

// NewMAFeatureMapper 新しいMAFeatureMapperを作成する
func NewMAFeatureMapper() *MAFeatureMapper {
	return &MAFeatureMapper{}
}

// ToFeatureVector 企業情報を特徴ベクトルに変換する
func (m *MAFeatureMapper) ToFeatureVector(
	company *domain.Company,
	financials []*domain.Financials,
	criteria *domain.MAMatchingCriteria,
) *matching.FeatureVector {
	fv := matching.NewFeatureVector(company.ID, "ma_company")

	if len(financials) == 0 {
		// 財務情報がない場合はメタデータのみ設定
		fv.SetMetadata("name", company.Name)
		fv.SetMetadata("purpose", string(company.MatchingPurpose))
		return fv
	}

	latest := financials[0] // 最新年度

	// 数値特徴（正規化）
	fv.SetNumerical("revenue", normalizeRevenue(latest.Revenue))
	fv.SetNumerical("ebitda_margin", normalizeMargin(latest.EBITDA, latest.Revenue))
	fv.SetNumerical("roe", normalizeROE(latest.ROE))
	fv.SetNumerical("roa", normalizeROA(latest.ROA))
	fv.SetNumerical("debt_equity_ratio", normalizeDebtEquity(latest.DebtEquityRatio))
	fv.SetNumerical("current_ratio", normalizeCurrentRatio(latest.CurrentRatio))
	fv.SetNumerical("employee_count", normalizeEmployees(company.EmployeeCount))

	// 時系列特徴（成長性）
	if len(financials) > 1 {
		revenueHistory := extractRevenue(financials)
		revenueStats := matching.ComputeTimeSeriesStats(revenueHistory)
		fv.SetTimeSeries("revenue_trend", revenueStats)

		ebitdaHistory := extractEBITDA(financials)
		ebitdaStats := matching.ComputeTimeSeriesStats(ebitdaHistory)
		fv.SetTimeSeries("ebitda_trend", ebitdaStats)
	}

	// カテゴリ特徴
	fv.SetCategorical("industry", string(company.Industry), 1.0)
	fv.SetCategorical("location", company.Location, 1.0)
	fv.SetCategorical("listing_status", string(company.ListingStatus), 1.0)
	fv.SetCategorical("stage", inferCompanyStage(company, latest), 1.0)

	// スパース特徴（技術ポートフォリオ、市場セグメント等）
	fv.EnsureSparse("technology")
	for _, tech := range company.Technologies {
		fv.SetSparse("technology", tech.Technology, 1.0)
	}

	fv.EnsureSparse("market")
	for _, market := range company.Markets {
		fv.SetSparse("market", market.Market, 1.0)
	}

	// メタデータ
	fv.SetMetadata("name", company.Name)
	fv.SetMetadata("purpose", string(company.MatchingPurpose))

	return fv
}

// normalizeRevenue 売上高を正規化する
// 1億円（8）〜1000億円（11）を0-1にマッピング（対数スケール）
func normalizeRevenue(revenue int64) float64 {
	if revenue <= 0 {
		return 0
	}
	logRevenue := math.Log10(float64(revenue))
	return matching.NormalizeValue(logRevenue, 8, 11)
}

// normalizeMargin EBITDAマージンを正規化する
// -50%〜+50%を0-1にマッピング
func normalizeMargin(ebitda, revenue int64) float64 {
	if revenue == 0 {
		return 0.5
	}
	margin := float64(ebitda) / float64(revenue)
	return matching.NormalizeValue(margin, -0.5, 0.5)
}

// normalizeROE 自己資本利益率を正規化する
// -50%〜+100%を0-1にマッピング
func normalizeROE(roe float64) float64 {
	return matching.NormalizeValue(roe, -0.5, 1.0)
}

// normalizeROA 総資産利益率を正規化する
// -30%〜+50%を0-1にマッピング
func normalizeROA(roa float64) float64 {
	return matching.NormalizeValue(roa, -0.3, 0.5)
}

// normalizeDebtEquity 負債比率を正規化する
// 0〜5倍を0-1にマッピング
func normalizeDebtEquity(ratio float64) float64 {
	return matching.NormalizeValue(ratio, 0, 5.0)
}

// normalizeCurrentRatio 流動比率を正規化する
// 0〜5倍を0-1にマッピング
func normalizeCurrentRatio(ratio float64) float64 {
	return matching.NormalizeValue(ratio, 0, 5.0)
}

// normalizeEmployees 従業員数を正規化する
// 1〜10000人を0-1にマッピング（対数スケール）
func normalizeEmployees(count int) float64 {
	if count <= 0 {
		return 0
	}
	logCount := math.Log10(float64(count))
	return matching.NormalizeValue(logCount, 0, 4)
}

// inferCompanyStage 企業ステージを推定する
func inferCompanyStage(company *domain.Company, financials *domain.Financials) string {
	age := time.Since(company.Founded).Hours() / 24 / 365

	if age < 5 && financials.Revenue < 1_000_000_000 {
		return string(domain.StageStartup)
	}
	if financials.ROE > 0.15 && financials.Revenue > 10_000_000_000 {
		return string(domain.StageMature)
	}
	if financials.ROE < 0 || financials.NetIncome < 0 {
		return string(domain.StageTurnaround)
	}
	return string(domain.StageGrowth)
}

// extractRevenue 財務情報から売上高の時系列を抽出する
func extractRevenue(financials []*domain.Financials) []float64 {
	result := make([]float64, len(financials))
	for i, f := range financials {
		result[i] = float64(f.Revenue)
	}
	return result
}

// extractEBITDA 財務情報からEBITDAの時系列を抽出する
func extractEBITDA(financials []*domain.Financials) []float64 {
	result := make([]float64, len(financials))
	for i, f := range financials {
		result[i] = float64(f.EBITDA)
	}
	return result
}
