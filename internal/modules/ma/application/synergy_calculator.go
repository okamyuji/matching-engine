package application

import (
	"github.com/okamyuji/matching-engine/internal/core/matching"
	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
)

// SynergyCalculator シナジー計算
type SynergyCalculator struct{}

// NewSynergyCalculator 新しいSynergyCalculatorを作成する
func NewSynergyCalculator() *SynergyCalculator {
	return &SynergyCalculator{}
}

// Calculate シナジータイプと期待値を計算する
func (c *SynergyCalculator) Calculate(
	source, candidate *matching.FeatureVector,
) *SynergySummary {
	// 1. 業界一致度でタイプ判定
	synergyType := c.determineSynergyType(source, candidate)

	// 2. 技術重複度（技術獲得のポテンシャル）
	technologyFit := c.calculateTechnologyFit(source, candidate)

	// 3. 顧客セグメント補完性
	customerFit := c.calculateCustomerFit(source, candidate)

	// 4. オペレーション適合度（業界・地域の類似性）
	operationalFit := c.calculateOperationalFit(source, candidate)

	// 5. 総合シナジー期待値
	expectedSynergy := c.calculateExpectedSynergy(
		synergyType,
		technologyFit,
		customerFit,
		operationalFit,
	)

	return &SynergySummary{
		Type:            synergyType,
		ExpectedSynergy: expectedSynergy,
		TechnologyFit:   technologyFit,
		CustomerFit:     customerFit,
		OperationalFit:  operationalFit,
	}
}

// determineSynergyType シナジータイプを判定する
func (c *SynergyCalculator) determineSynergyType(source, candidate *matching.FeatureVector) string {
	sourceIndustry := c.getCategoricalValue(source, "industry")
	candidateIndustry := c.getCategoricalValue(candidate, "industry")

	// 技術ポートフォリオの重複度
	techOverlap := c.calculateSparseOverlap(source, candidate, "technology")

	// 同一業界で技術重複が低い場合: 水平統合
	if sourceIndustry == candidateIndustry {
		if techOverlap < 0.3 {
			return string(domain.SynergyTechnology)
		}
		return string(domain.SynergyHorizontal)
	}

	// 異業界で技術重複が高い場合: 垂直統合の可能性
	if techOverlap > 0.3 {
		return string(domain.SynergyVertical)
	}

	// その他: 多角化
	return string(domain.SynergyDiversification)
}

// calculateTechnologyFit 技術適合度を計算する
func (c *SynergyCalculator) calculateTechnologyFit(source, candidate *matching.FeatureVector) float64 {
	// 技術の補完性（重複が少ないほど高スコア）
	overlap := c.calculateSparseOverlap(source, candidate, "technology")

	sourceTechs := c.getSparseFeatures(source, "technology")
	candidateTechs := c.getSparseFeatures(candidate, "technology")

	// 両社の技術数
	totalTechs := float64(len(sourceTechs) + len(candidateTechs))
	if totalTechs == 0 {
		return 0.5
	}

	// 補完性スコア（重複が少なく、両社が技術を持っている場合に高い）
	complementarity := 1.0 - overlap
	techRichness := totalTechs / 20.0 // 合計20技術で最大スコア
	if techRichness > 1.0 {
		techRichness = 1.0
	}

	return (complementarity*0.6 + techRichness*0.4)
}

// calculateCustomerFit 顧客適合度を計算する
func (c *SynergyCalculator) calculateCustomerFit(source, candidate *matching.FeatureVector) float64 {
	// 市場セグメントの補完性
	marketOverlap := c.calculateSparseOverlap(source, candidate, "market")

	// 顧客基盤の補完性（重複が少ないほど新規顧客獲得のポテンシャル）
	return 1.0 - marketOverlap
}

// calculateOperationalFit オペレーション適合度を計算する
func (c *SynergyCalculator) calculateOperationalFit(source, candidate *matching.FeatureVector) float64 {
	// 業界の類似性
	industrySame := 0.0
	if c.getCategoricalValue(source, "industry") == c.getCategoricalValue(candidate, "industry") {
		industrySame = 1.0
	}

	// 地域の類似性
	locationSame := 0.0
	if c.getCategoricalValue(source, "location") == c.getCategoricalValue(candidate, "location") {
		locationSame = 1.0
	}

	// 企業ステージの類似性
	stageSame := 0.0
	if c.getCategoricalValue(source, "stage") == c.getCategoricalValue(candidate, "stage") {
		stageSame = 1.0
	}

	// 加重平均
	return industrySame*0.5 + locationSame*0.3 + stageSame*0.2
}

// calculateExpectedSynergy 総合シナジー期待値を計算する
func (c *SynergyCalculator) calculateExpectedSynergy(
	synergyType string,
	technologyFit, customerFit, operationalFit float64,
) float64 {
	// シナジータイプごとに重み付け
	switch synergyType {
	case string(domain.SynergyHorizontal):
		// 水平統合: オペレーション効率とコスト削減
		return operationalFit*0.6 + customerFit*0.3 + technologyFit*0.1
	case string(domain.SynergyVertical):
		// 垂直統合: サプライチェーン最適化
		return technologyFit*0.5 + operationalFit*0.3 + customerFit*0.2
	case string(domain.SynergyTechnology):
		// 技術獲得: 技術ポートフォリオ強化
		return technologyFit*0.7 + customerFit*0.2 + operationalFit*0.1
	case string(domain.SynergyDiversification):
		// 多角化: リスク分散と新市場開拓
		return customerFit*0.5 + technologyFit*0.3 + operationalFit*0.2
	default:
		// デフォルト: 均等
		return (technologyFit + customerFit + operationalFit) / 3.0
	}
}

// getCategoricalValue カテゴリ特徴の値を取得する
func (c *SynergyCalculator) getCategoricalValue(fv *matching.FeatureVector, field string) string {
	if fv.Categorical == nil {
		return ""
	}
	cat := fv.Categorical[field]
	if cat == nil {
		return ""
	}
	for value := range cat {
		return value // 最初の値を返す（one-hotなので1つだけ）
	}
	return ""
}

// getSparseFeatures スパース特徴のリストを取得する
func (c *SynergyCalculator) getSparseFeatures(fv *matching.FeatureVector, field string) map[string]float64 {
	if fv.Sparse == nil {
		return make(map[string]float64)
	}
	sparse := fv.Sparse[field]
	if sparse == nil {
		return make(map[string]float64)
	}
	return sparse
}

// calculateSparseOverlap スパース特徴の重複度を計算する（Jaccard係数）
func (c *SynergyCalculator) calculateSparseOverlap(source, candidate *matching.FeatureVector, field string) float64 {
	sourceFeatures := c.getSparseFeatures(source, field)
	candidateFeatures := c.getSparseFeatures(candidate, field)

	if len(sourceFeatures) == 0 && len(candidateFeatures) == 0 {
		return 0.0
	}

	// 積集合を計算
	intersection := 0
	for key := range sourceFeatures {
		if _, exists := candidateFeatures[key]; exists {
			intersection++
		}
	}

	// 和集合を計算
	union := len(sourceFeatures) + len(candidateFeatures) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
