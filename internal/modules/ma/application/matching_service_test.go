package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/core/matching"
	"github.com/okamyuji/matching-engine/internal/modules/ma/domain"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/mapper"
	"github.com/okamyuji/matching-engine/internal/modules/ma/infrastructure/repository"
)

// モックリポジトリ実装

type mockCompanyRepository struct {
	companies  map[string]*domain.Company
	criteria   map[string]*domain.MAMatchingCriteria
	candidates []*repository.CompanyWithFinancials
	err        error
}

func newMockCompanyRepository() *mockCompanyRepository {
	return &mockCompanyRepository{
		companies:  make(map[string]*domain.Company),
		criteria:   make(map[string]*domain.MAMatchingCriteria),
		candidates: make([]*repository.CompanyWithFinancials, 0),
	}
}

func (m *mockCompanyRepository) FindByID(ctx context.Context, id string) (*domain.Company, error) {
	if m.err != nil {
		return nil, m.err
	}
	if c, ok := m.companies[id]; ok {
		return c, nil
	}
	return nil, errors.New("company not found")
}

func (m *mockCompanyRepository) FindByPurpose(ctx context.Context, purpose domain.MatchingPurpose, criteria *domain.MAMatchingCriteria) ([]*repository.CompanyWithFinancials, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.candidates, nil
}

func (m *mockCompanyRepository) FindCriteria(ctx context.Context, companyID string) (*domain.MAMatchingCriteria, error) {
	if m.err != nil {
		return nil, m.err
	}
	if c, ok := m.criteria[companyID]; ok {
		return c, nil
	}
	return nil, errors.New("criteria not found")
}

func (m *mockCompanyRepository) Create(ctx context.Context, company *domain.Company) error {
	return m.err
}

func (m *mockCompanyRepository) Update(ctx context.Context, company *domain.Company) error {
	return m.err
}

type mockFinancialsRepository struct {
	financials map[string][]*domain.Financials
	err        error
}

func newMockFinancialsRepository() *mockFinancialsRepository {
	return &mockFinancialsRepository{
		financials: make(map[string][]*domain.Financials),
	}
}

func (m *mockFinancialsRepository) FindByCompanyID(ctx context.Context, companyID string, years int) ([]*domain.Financials, error) {
	if m.err != nil {
		return nil, m.err
	}
	if f, ok := m.financials[companyID]; ok {
		return f, nil
	}
	return []*domain.Financials{}, nil
}

func (m *mockFinancialsRepository) FindLatest(ctx context.Context, companyID string) (*domain.Financials, error) {
	if m.err != nil {
		return nil, m.err
	}
	if f, ok := m.financials[companyID]; ok && len(f) > 0 {
		return f[0], nil
	}
	return nil, errors.New("financials not found")
}

func (m *mockFinancialsRepository) Save(ctx context.Context, financials *domain.Financials) error {
	return m.err
}

type mockInterestRepository struct {
	interests []*domain.Interest
	existsMap map[string]bool // key: "fromID:toID"
	err       error
}

func newMockInterestRepository() *mockInterestRepository {
	return &mockInterestRepository{
		interests: make([]*domain.Interest, 0),
		existsMap: make(map[string]bool),
	}
}

func (m *mockInterestRepository) Save(ctx context.Context, interest *domain.Interest) error {
	if m.err != nil {
		return m.err
	}
	m.interests = append(m.interests, interest)
	return nil
}

func (m *mockInterestRepository) FindByToCompany(ctx context.Context, toCompanyID string) ([]*domain.Interest, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*domain.Interest, 0)
	for _, i := range m.interests {
		if i.ToCompanyID == toCompanyID {
			result = append(result, i)
		}
	}
	return result, nil
}

func (m *mockInterestRepository) Exists(ctx context.Context, fromCompanyID, toCompanyID string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	key := fromCompanyID + ":" + toCompanyID
	return m.existsMap[key], nil
}

type mockMAMatchRepository struct {
	matches []*domain.MAMatch
	err     error
}

func newMockMAMatchRepository() *mockMAMatchRepository {
	return &mockMAMatchRepository{
		matches: make([]*domain.MAMatch, 0),
	}
}

func (m *mockMAMatchRepository) Save(ctx context.Context, match *domain.MAMatch) error {
	if m.err != nil {
		return m.err
	}
	m.matches = append(m.matches, match)
	return nil
}

func (m *mockMAMatchRepository) FindMutual(ctx context.Context, companyID string) ([]*domain.MAMatch, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]*domain.MAMatch, 0)
	for _, m := range m.matches {
		if m.CompanyIDA == companyID || m.CompanyIDB == companyID {
			result = append(result, m)
		}
	}
	return result, nil
}

func (m *mockMAMatchRepository) FindByID(ctx context.Context, matchID string) (*domain.MAMatch, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, match := range m.matches {
		if match.ID == matchID {
			return match, nil
		}
	}
	return nil, errors.New("match not found")
}

// テストヘルパー

func createTestEngine() *matching.ConfigurableEngine {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "revenue",
					Type:   "euclidean",
					Fields: []string{"revenue"},
					Weight: 1.0,
				},
			},
		},
		Ranking: matching.RankingConfig{
			SortOrder: "desc",
		},
	}

	engine, err := matching.NewConfigurableEngine(config)
	if err != nil {
		panic(err)
	}
	return engine
}

// FindTargetsテスト

func TestMAMatchingService_FindTargets_Success(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	companyRepo.companies["buyer-1"] = &domain.Company{
		ID:              "buyer-1",
		Name:            "Buyer Company",
		Industry:        domain.IndustryTechnology,
		MatchingPurpose: domain.PurposeBuyer,
	}
	companyRepo.criteria["buyer-1"] = &domain.MAMatchingCriteria{
		CompanyID: "buyer-1",
		Purpose:   domain.PurposeBuyer,
	}
	companyRepo.candidates = []*repository.CompanyWithFinancials{
		{
			Company: &domain.Company{
				ID:              "seller-1",
				Name:            "Seller Company",
				Industry:        domain.IndustryTechnology,
				MatchingPurpose: domain.PurposeSeller,
			},
			Financials: []*domain.Financials{
				{
					CompanyID:       "seller-1",
					FiscalYear:      2023,
					Revenue:         1000000000,
					EBITDA:          200000000,
					NetIncome:       100000000,
					ROE:             0.15,
					ROA:             0.08,
					DebtEquityRatio: 1.5,
				},
			},
		},
	}

	financialsRepo := newMockFinancialsRepository()
	financialsRepo.financials["buyer-1"] = []*domain.Financials{
		{
			CompanyID:  "buyer-1",
			FiscalYear: 2023,
			Revenue:    2000000000,
			EBITDA:     400000000,
			NetIncome:  200000000,
		},
	}

	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	results, err := service.FindTargets(ctx, "buyer-1", 10)
	if err != nil {
		t.Fatalf("FindTargets failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	if len(results) > 0 {
		if results[0].CompanyID != "seller-1" {
			t.Errorf("Expected seller-1, got %s", results[0].CompanyID)
		}
	}
}

func TestMAMatchingService_FindTargets_NoCandidates(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	companyRepo.companies["buyer-1"] = &domain.Company{
		ID:              "buyer-1",
		Name:            "Buyer Company",
		Industry:        domain.IndustryTechnology,
		MatchingPurpose: domain.PurposeBuyer,
	}
	companyRepo.candidates = []*repository.CompanyWithFinancials{} // 候補なし

	financialsRepo := newMockFinancialsRepository()
	financialsRepo.financials["buyer-1"] = []*domain.Financials{
		{
			CompanyID:  "buyer-1",
			FiscalYear: 2023,
			Revenue:    2000000000,
		},
	}

	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	results, err := service.FindTargets(ctx, "buyer-1", 10)
	if err != nil {
		t.Fatalf("FindTargets failed: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %d", len(results))
	}
}

func TestMAMatchingService_FindTargets_CompanyNotFound(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	_, err := service.FindTargets(ctx, "nonexistent", 10)
	if err == nil {
		t.Error("Expected error for nonexistent company")
	}
}

func TestMAMatchingService_FindTargets_FinancialsError(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	companyRepo.companies["buyer-1"] = &domain.Company{
		ID:              "buyer-1",
		Name:            "Buyer Company",
		MatchingPurpose: domain.PurposeBuyer,
	}

	financialsRepo := newMockFinancialsRepository()
	financialsRepo.err = errors.New("database error")

	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	_, err := service.FindTargets(ctx, "buyer-1", 10)
	if err == nil {
		t.Error("Expected error when financials fetch fails")
	}
}

func TestMAMatchingService_FindTargets_NoCriteria(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	companyRepo.companies["buyer-1"] = &domain.Company{
		ID:              "buyer-1",
		Name:            "Buyer Company",
		Industry:        domain.IndustryTechnology,
		MatchingPurpose: domain.PurposeBuyer,
	}
	// criteriaは設定しない（FindCriteriaがエラーを返す）
	companyRepo.candidates = []*repository.CompanyWithFinancials{}

	financialsRepo := newMockFinancialsRepository()
	financialsRepo.financials["buyer-1"] = []*domain.Financials{
		{
			CompanyID:  "buyer-1",
			FiscalYear: 2023,
			Revenue:    2000000000,
		},
	}

	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	// criteriaがない場合でもデフォルトが使われるので成功する
	_, err := service.FindTargets(ctx, "buyer-1", 10)
	if err != nil {
		t.Fatalf("FindTargets should succeed with default criteria: %v", err)
	}
}

// SendInterestテスト

func TestMAMatchingService_SendInterest_Success_NoMatch(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	response, err := service.SendInterest(ctx, "buyer-1", "seller-1")
	if err != nil {
		t.Fatalf("SendInterest failed: %v", err)
	}

	if response.Matched {
		t.Error("Expected Matched=false when no reverse interest exists")
	}

	if response.MatchID != "" {
		t.Error("Expected empty MatchID when no match")
	}

	if len(interestRepo.interests) != 1 {
		t.Errorf("Expected 1 interest saved, got %d", len(interestRepo.interests))
	}
}

func TestMAMatchingService_SendInterest_Success_WithMatch(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	// 相手からの興味表明が既に存在
	interestRepo.existsMap["seller-1:buyer-1"] = true

	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	response, err := service.SendInterest(ctx, "buyer-1", "seller-1")
	if err != nil {
		t.Fatalf("SendInterest failed: %v", err)
	}

	if !response.Matched {
		t.Error("Expected Matched=true when reverse interest exists")
	}

	if response.MatchID == "" {
		t.Error("Expected non-empty MatchID when match occurs")
	}

	if len(matchRepo.matches) != 1 {
		t.Errorf("Expected 1 match saved, got %d", len(matchRepo.matches))
	}
}

func TestMAMatchingService_SendInterest_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	// 既に興味表明済み
	interestRepo.existsMap["buyer-1:seller-1"] = true

	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	_, err := service.SendInterest(ctx, "buyer-1", "seller-1")
	if err == nil {
		t.Error("Expected error when interest already exists")
	}
}

func TestMAMatchingService_SendInterest_SaveError(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	interestRepo.err = errors.New("database error")

	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	_, err := service.SendInterest(ctx, "buyer-1", "seller-1")
	if err == nil {
		t.Error("Expected error when interest save fails")
	}
}

// GetReceivedInterestsテスト

func TestMAMatchingService_GetReceivedInterests_Success(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	interestRepo.interests = []*domain.Interest{
		{
			ID:            "interest-1",
			FromCompanyID: "buyer-1",
			ToCompanyID:   "seller-1",
			CreatedAt:     time.Now(),
		},
		{
			ID:            "interest-2",
			FromCompanyID: "buyer-2",
			ToCompanyID:   "seller-1",
			CreatedAt:     time.Now(),
		},
	}

	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	interests, err := service.GetReceivedInterests(ctx, "seller-1")
	if err != nil {
		t.Fatalf("GetReceivedInterests failed: %v", err)
	}

	if len(interests) != 2 {
		t.Errorf("Expected 2 interests, got %d", len(interests))
	}
}

func TestMAMatchingService_GetReceivedInterests_Empty(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	interests, err := service.GetReceivedInterests(ctx, "seller-1")
	if err != nil {
		t.Fatalf("GetReceivedInterests failed: %v", err)
	}

	if len(interests) != 0 {
		t.Errorf("Expected 0 interests, got %d", len(interests))
	}
}

func TestMAMatchingService_GetReceivedInterests_Error(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	interestRepo.err = errors.New("database error")

	matchRepo := newMockMAMatchRepository()
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	_, err := service.GetReceivedInterests(ctx, "seller-1")
	if err == nil {
		t.Error("Expected error when interest fetch fails")
	}
}

// GetMutualMatchesテスト

func TestMAMatchingService_GetMutualMatches_Success(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	matchRepo.matches = []*domain.MAMatch{
		{
			ID:         "match-1",
			CompanyIDA: "buyer-1",
			CompanyIDB: "seller-1",
			Score:      0.85,
			MatchedAt:  time.Now(),
		},
		{
			ID:         "match-2",
			CompanyIDA: "buyer-1",
			CompanyIDB: "seller-2",
			Score:      0.75,
			MatchedAt:  time.Now(),
		},
	}

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	matches, err := service.GetMutualMatches(ctx, "buyer-1")
	if err != nil {
		t.Fatalf("GetMutualMatches failed: %v", err)
	}

	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}
}

func TestMAMatchingService_GetMutualMatches_Empty(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	matches, err := service.GetMutualMatches(ctx, "buyer-1")
	if err != nil {
		t.Fatalf("GetMutualMatches failed: %v", err)
	}

	if len(matches) != 0 {
		t.Errorf("Expected 0 matches, got %d", len(matches))
	}
}

func TestMAMatchingService_GetMutualMatches_Error(t *testing.T) {
	ctx := context.Background()
	engine := createTestEngine()

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	matchRepo.err = errors.New("database error")

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := NewSynergyCalculator()

	service := NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	_, err := service.GetMutualMatches(ctx, "buyer-1")
	if err == nil {
		t.Error("Expected error when match fetch fails")
	}
}
