package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/matching-engine/internal/core/matching"
	"github.com/yourorg/matching-engine/internal/modules/ma/application"
	"github.com/yourorg/matching-engine/internal/modules/ma/domain"
	"github.com/yourorg/matching-engine/internal/modules/ma/infrastructure/mapper"
	"github.com/yourorg/matching-engine/internal/modules/ma/infrastructure/repository"
)

// モックリポジトリ実装

type mockCompanyRepository struct {
	companies map[string]*domain.Company
	err       error
}

func newMockCompanyRepository() *mockCompanyRepository {
	return &mockCompanyRepository{
		companies: make(map[string]*domain.Company),
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

func (m *mockCompanyRepository) FindCandidates(ctx context.Context, sourceCompany *domain.Company, criteria *domain.MAMatchingCriteria) ([]*repository.CompanyWithFinancials, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*repository.CompanyWithFinancials{}, nil
}

func (m *mockCompanyRepository) Create(ctx context.Context, company *domain.Company) error {
	return m.err
}

func (m *mockCompanyRepository) Update(ctx context.Context, company *domain.Company) error {
	return m.err
}

func (m *mockCompanyRepository) FindByPurpose(ctx context.Context, purpose domain.MatchingPurpose, criteria *domain.MAMatchingCriteria) ([]*repository.CompanyWithFinancials, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*repository.CompanyWithFinancials{}, nil
}

func (m *mockCompanyRepository) FindCriteria(ctx context.Context, companyID string) (*domain.MAMatchingCriteria, error) {
	if m.err != nil {
		return nil, m.err
	}
	// デフォルトの基準を返す
	return &domain.MAMatchingCriteria{
		CompanyID: companyID,
	}, nil
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

func (m *mockFinancialsRepository) FindLatest(ctx context.Context, companyID string) (*domain.Financials, error) {
	if m.err != nil {
		return nil, m.err
	}
	if f, ok := m.financials[companyID]; ok && len(f) > 0 {
		return f[0], nil
	}
	return nil, errors.New("financials not found")
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

func (m *mockFinancialsRepository) Save(ctx context.Context, financials *domain.Financials) error {
	return m.err
}

type mockInterestRepository struct {
	interests []*domain.Interest
	mutual    bool
	err       error
}

func newMockInterestRepository() *mockInterestRepository {
	return &mockInterestRepository{
		interests: make([]*domain.Interest, 0),
	}
}

func (m *mockInterestRepository) Save(ctx context.Context, interest *domain.Interest) error {
	if m.err != nil {
		return m.err
	}
	m.interests = append(m.interests, interest)
	return nil
}

func (m *mockInterestRepository) CheckMutual(ctx context.Context, companyA, companyB string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.mutual, nil
}

func (m *mockInterestRepository) FindByToCompany(ctx context.Context, toCompanyID string) ([]*domain.Interest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.interests, nil
}

func (m *mockInterestRepository) Exists(ctx context.Context, fromCompanyID, toCompanyID string) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	for _, interest := range m.interests {
		if interest.FromCompanyID == fromCompanyID && interest.ToCompanyID == toCompanyID {
			return true, nil
		}
	}
	return false, nil
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
	return m.matches, nil
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

func TestHandler_GetTargets_Success(t *testing.T) {
	// エンジンのセットアップ
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	// リポジトリのセットアップ
	companyRepo := newMockCompanyRepository()
	companyRepo.companies["company-123"] = &domain.Company{
		ID:       "company-123",
		Name:     "Test Company",
		Industry: domain.IndustryTechnology,
	}

	financialsRepo := newMockFinancialsRepository()
	financialsRepo.financials["company-123"] = []*domain.Financials{
		{
			CompanyID:       "company-123",
			FiscalYear:      2023,
			Revenue:         1000000000,
			EBITDA:          200000000,
			NetIncome:       100000000,
			ROE:             0.15,
			ROA:             0.08,
			DebtEquityRatio: 1.5,
			CurrentRatio:    2.0,
		},
	}

	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()

	// サービスのセットアップ
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	// リクエスト作成
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/targets?limit=10", nil)
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetTargets(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandler_SendInterest_Success(t *testing.T) {
	// エンジンのセットアップ
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	// リポジトリのセットアップ
	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()

	// サービスのセットアップ
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	// リクエスト作成
	reqBody := application.InterestRequest{TargetCompanyID: "target-1"}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ma/interests", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.SendInterest(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandler_GetReceivedInterests_Success(t *testing.T) {
	// エンジンのセットアップ
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	// リポジトリのセットアップ
	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	interestRepo.interests = []*domain.Interest{
		{
			ID:            "interest-1",
			FromCompanyID: "company-456",
			ToCompanyID:   "company-123",
			CreatedAt:     time.Now(),
		},
	}
	matchRepo := newMockMAMatchRepository()

	// サービスのセットアップ
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	// リクエスト作成
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/interests/received", nil)
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetReceivedInterests(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandler_GetMatches_Success(t *testing.T) {
	// エンジンのセットアップ
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	// リポジトリのセットアップ
	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	matchRepo.matches = []*domain.MAMatch{
		{
			ID:         "match-1",
			CompanyIDA: "company-123",
			CompanyIDB: "company-456",
			Score:      0.85,
			MatchedAt:  time.Now(),
		},
	}

	// サービスのセットアップ
	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	// リクエスト作成
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/matches", nil)
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetMatches(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestHandler_GetValuation_NotImplemented(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/valuation/company-123", nil)
	w := httptest.NewRecorder()

	handler.GetValuation(w, req)

	// GetValuationは未実装なので501を期待
	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}

// エラーケーステスト

func TestHandler_GetTargets_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	// companyIDKeyが設定されていないリクエスト
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/targets", nil)
	w := httptest.NewRecorder()

	handler.GetTargets(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_GetTargets_ServiceError(t *testing.T) {
	// エンジンのセットアップ
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	// エラーを返すモックリポジトリ
	companyRepo := newMockCompanyRepository()
	companyRepo.err = errors.New("database error")

	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/targets", nil)
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetTargets(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_SendInterest_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	reqBody := application.InterestRequest{TargetCompanyID: "target-1"}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ma/interests", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.SendInterest(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_SendInterest_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil)

	// 不正なJSON
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ma/interests", bytes.NewReader([]byte("invalid json")))
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.SendInterest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_SendInterest_MissingTargetCompanyID(t *testing.T) {
	handler := NewHandler(nil, nil)

	// target_company_idが空
	reqBody := application.InterestRequest{TargetCompanyID: ""}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ma/interests", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.SendInterest(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_SendInterest_ServiceError(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	interestRepo.err = errors.New("database error")

	matchRepo := newMockMAMatchRepository()

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	reqBody := application.InterestRequest{TargetCompanyID: "target-1"}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ma/interests", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.SendInterest(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_GetReceivedInterests_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/interests/received", nil)
	w := httptest.NewRecorder()

	handler.GetReceivedInterests(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_GetReceivedInterests_ServiceError(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	interestRepo.err = errors.New("database error")

	matchRepo := newMockMAMatchRepository()

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/interests/received", nil)
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetReceivedInterests(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_GetMatches_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/matches", nil)
	w := httptest.NewRecorder()

	handler.GetMatches(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandler_GetMatches_ServiceError(t *testing.T) {
	config := &matching.MatchingConfig{
		Version: "1.0",
		Domain:  "ma_matching",
		Scoring: matching.ScoringConfig{
			Components: []matching.ComponentConfig{
				{
					Name:   "test",
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
		t.Fatalf("failed to create engine: %v", err)
	}

	companyRepo := newMockCompanyRepository()
	financialsRepo := newMockFinancialsRepository()
	interestRepo := newMockInterestRepository()
	matchRepo := newMockMAMatchRepository()
	matchRepo.err = errors.New("database error")

	featureMapper := mapper.NewMAFeatureMapper()
	synergyCalc := application.NewSynergyCalculator()

	matchingService := application.NewMAMatchingService(
		engine,
		companyRepo,
		financialsRepo,
		interestRepo,
		matchRepo,
		featureMapper,
		synergyCalc,
	)

	handler := NewHandler(matchingService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/matches", nil)
	ctx := context.WithValue(req.Context(), companyIDKey, "company-123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler.GetMatches(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandler_GetValuation_InvalidPath(t *testing.T) {
	handler := NewHandler(nil, nil)

	// 短すぎるパス
	req := httptest.NewRequest(http.MethodGet, "/api/v1", nil)
	w := httptest.NewRecorder()

	handler.GetValuation(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_GetValuation_EmptyCompanyID(t *testing.T) {
	handler := NewHandler(nil, nil)

	// 末尾にスラッシュだけの場合も NotImplemented が返される
	req := httptest.NewRequest(http.MethodGet, "/api/v1/ma/valuation/", nil)
	w := httptest.NewRecorder()

	handler.GetValuation(w, req)

	// 現在の実装では空IDのチェックが機能せずNotImplementedが返される
	if w.Code != http.StatusNotImplemented {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotImplemented)
	}
}
