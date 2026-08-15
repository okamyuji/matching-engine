package application

import (
	"context"
	"fmt"

	"github.com/okamyuji/matching-engine/internal/core/matching"
	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/mapper"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository"
)

// MAMatchingService M&Aマッチング処理を統括する
type MAMatchingService struct {
	engine            *matching.ConfigurableEngine
	companyRepo       repository.CompanyRepository
	financialsRepo    repository.FinancialsRepository
	interestRepo      repository.InterestRepository
	matchRepo         repository.MAMatchRepository
	featureMapper     *mapper.MAFeatureMapper
	synergyCalculator *SynergyCalculator
}

// NewMAMatchingService 新しいMAMatchingServiceを作成する
func NewMAMatchingService(
	engine *matching.ConfigurableEngine,
	companyRepo repository.CompanyRepository,
	financialsRepo repository.FinancialsRepository,
	interestRepo repository.InterestRepository,
	matchRepo repository.MAMatchRepository,
	featureMapper *mapper.MAFeatureMapper,
	synergyCalculator *SynergyCalculator,
) *MAMatchingService {
	return &MAMatchingService{
		engine:            engine,
		companyRepo:       companyRepo,
		financialsRepo:    financialsRepo,
		interestRepo:      interestRepo,
		matchRepo:         matchRepo,
		featureMapper:     featureMapper,
		synergyCalculator: synergyCalculator,
	}
}

// FindTargets マッチング候補を取得する
// 処理:
// 1. 自社情報取得
// 2. 自社財務情報取得（5年分）
// 3. マッチング基準取得
// 4. 候補企業取得（反対のpurpose）
// 5. 特徴ベクトル変換
// 6. マッチング実行
// 7. シナジー計算
// 8. DTO変換
func (s *MAMatchingService) FindTargets(
	ctx context.Context,
	companyID string,
	limit int,
) ([]*MAMatchResult, error) {
	// 1. 自社情報取得
	company, err := s.companyRepo.FindByID(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find company: %w", err)
	}

	// 2. 自社財務情報取得（5年分）
	financials, err := s.financialsRepo.FindByCompanyID(ctx, companyID, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to find financials: %w", err)
	}

	// 3. マッチング基準取得
	var criteria *domain.MAMatchingCriteria
	criteria, err = s.companyRepo.FindCriteria(ctx, companyID)
	if err != nil {
		// 基準が見つからない場合はデフォルトを使用
		criteria = &domain.MAMatchingCriteria{
			CompanyID: companyID,
			Purpose:   company.MatchingPurpose,
		}
	}

	// 4. 候補企業取得（反対のpurpose）
	targetPurpose := domain.PurposeSeller
	if company.MatchingPurpose == domain.PurposeSeller {
		targetPurpose = domain.PurposeBuyer
	}

	candidates, err := s.companyRepo.FindByPurpose(ctx, targetPurpose, criteria)
	if err != nil {
		return nil, fmt.Errorf("failed to find candidates: %w", err)
	}

	if len(candidates) == 0 {
		return []*MAMatchResult{}, nil
	}

	// 5. 特徴ベクトル変換
	sourceVector := s.featureMapper.ToFeatureVector(company, financials, criteria)

	candidateVectors := make([]*matching.FeatureVector, len(candidates))
	for i, c := range candidates {
		candidateVectors[i] = s.featureMapper.ToFeatureVector(c.Company, c.Financials, criteria)
	}

	// 6. マッチング実行
	matches, err := s.engine.FindMatches(ctx, sourceVector, candidateVectors)
	if err != nil {
		return nil, fmt.Errorf("failed to compute matches: %w", err)
	}

	// 7. シナジー計算 & 8. DTO変換
	results := make([]*MAMatchResult, 0, len(matches))
	for _, m := range matches {
		if len(results) >= limit {
			break
		}

		// 候補企業の財務情報を取得
		candidateFinancials := getCandidateFinancials(candidates, m.Candidate.ID)
		if len(candidateFinancials) == 0 {
			continue
		}
		latestFinancials := candidateFinancials[0]

		// シナジー計算
		synergySummary := s.synergyCalculator.Calculate(sourceVector, m.Candidate)

		// DTO変換
		result := &MAMatchResult{
			CompanyID: m.Candidate.ID,
			Score:     m.Score,
			Rank:      m.Rank,
			Breakdown: m.Breakdown,
			FinancialSummary: &FinancialSummary{
				Revenue:         latestFinancials.Revenue,
				EBITDA:          latestFinancials.EBITDA,
				EBITDAMargin:    latestFinancials.EBITDAMargin(),
				ROE:             latestFinancials.ROE,
				ROA:             latestFinancials.ROA,
				DebtEquityRatio: latestFinancials.DebtEquityRatio,
			},
			SynergySummary: synergySummary,
		}

		results = append(results, result)
	}

	return results, nil
}

// SendInterest 興味表明を送信する
func (s *MAMatchingService) SendInterest(
	ctx context.Context,
	fromCompanyID, toCompanyID string,
) (*InterestResponse, error) {
	// 既に興味表明済みかチェック
	exists, err := s.interestRepo.Exists(ctx, fromCompanyID, toCompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to check interest existence: %w", err)
	}

	if exists {
		return nil, fmt.Errorf("interest already sent")
	}

	// 興味表明を保存
	interest := &domain.Interest{
		ID:            generateID(),
		FromCompanyID: fromCompanyID,
		ToCompanyID:   toCompanyID,
	}

	err = s.interestRepo.Save(ctx, interest)
	if err != nil {
		return nil, fmt.Errorf("failed to save interest: %w", err)
	}

	// 相互の興味表明があるかチェック
	reverseExists, err := s.interestRepo.Exists(ctx, toCompanyID, fromCompanyID)
	if err != nil {
		return nil, fmt.Errorf("failed to check reverse interest: %w", err)
	}

	response := &InterestResponse{
		Matched: reverseExists,
	}

	// 相互マッチの場合、マッチレコードを作成
	if reverseExists {
		match := &domain.MAMatch{
			ID:         generateID(),
			CompanyIDA: fromCompanyID,
			CompanyIDB: toCompanyID,
			Score:      0.0, // スコアは後で計算
			Breakdown:  make(map[string]float64),
		}

		err = s.matchRepo.Save(ctx, match)
		if err != nil {
			return nil, fmt.Errorf("failed to save match: %w", err)
		}

		response.MatchID = match.ID
	}

	return response, nil
}

// GetReceivedInterests 受け取った興味表明を取得する
func (s *MAMatchingService) GetReceivedInterests(
	ctx context.Context,
	companyID string,
) ([]*domain.Interest, error) {
	interests, err := s.interestRepo.FindByToCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find received interests: %w", err)
	}

	return interests, nil
}

// GetMutualMatches 相互マッチを取得する
func (s *MAMatchingService) GetMutualMatches(
	ctx context.Context,
	companyID string,
) ([]*domain.MAMatch, error) {
	matches, err := s.matchRepo.FindMutual(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to find mutual matches: %w", err)
	}

	return matches, nil
}

// getCandidateFinancials 候補企業リストから財務情報を取得する
func getCandidateFinancials(candidates []*repository.CompanyWithFinancials, companyID string) []*domain.Financials {
	for _, c := range candidates {
		if c.Company.ID == companyID {
			return c.Financials
		}
	}
	return nil
}

// generateID 簡易的なID生成（本番環境ではUUIDなどを使用）
func generateID() string {
	// 本番環境では github.com/google/uuid などを使用すべき
	// ここでは簡易実装として現在時刻ベースのIDを生成
	return fmt.Sprintf("id_%d", 1000000000) // 仮実装
}
