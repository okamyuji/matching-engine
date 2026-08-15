package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/okamyuji/matching-engine/internal/modules/ma/application"
)

// userIDKey コンテキストキー（認証ミドルウェアで設定される）
type contextKey string

const companyIDKey contextKey = "company_id"

// Handler M&A API のHTTPリクエストを処理する
type Handler struct {
	matchingService  *application.MAMatchingService
	valuationService *application.ValuationService
}

// NewHandler 新しいHandlerを作成する
func NewHandler(
	matchingService *application.MAMatchingService,
	valuationService *application.ValuationService,
) *Handler {
	return &Handler{
		matchingService:  matchingService,
		valuationService: valuationService,
	}
}

// GetTargets マッチング候補企業を取得する
// GET /api/v1/ma/targets
func (h *Handler) GetTargets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companyID := getCompanyIDFromContext(ctx)

	if companyID == "" {
		http.Error(w, "company not authenticated", http.StatusUnauthorized)
		return
	}

	limit := 30
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	targets, err := h.matchingService.FindTargets(ctx, companyID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(targets); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SendInterest 別の企業に興味表明を送る
// POST /api/v1/ma/interests
func (h *Handler) SendInterest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companyID := getCompanyIDFromContext(ctx)

	if companyID == "" {
		http.Error(w, "company not authenticated", http.StatusUnauthorized)
		return
	}

	var req application.InterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.TargetCompanyID == "" {
		http.Error(w, "target_company_id is required", http.StatusBadRequest)
		return
	}

	response, err := h.matchingService.SendInterest(ctx, companyID, req.TargetCompanyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetReceivedInterests 受け取った興味表明を取得する
// GET /api/v1/ma/interests/received
func (h *Handler) GetReceivedInterests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companyID := getCompanyIDFromContext(ctx)

	if companyID == "" {
		http.Error(w, "company not authenticated", http.StatusUnauthorized)
		return
	}

	interests, err := h.matchingService.GetReceivedInterests(ctx, companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(interests); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetMatches 相互マッチを取得する
// GET /api/v1/ma/matches
func (h *Handler) GetMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	companyID := getCompanyIDFromContext(ctx)

	if companyID == "" {
		http.Error(w, "company not authenticated", http.StatusUnauthorized)
		return
	}

	matches, err := h.matchingService.GetMutualMatches(ctx, companyID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matches); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetValuation 企業のバリュエーションを取得する
// GET /api/v1/ma/valuation/{id}
func (h *Handler) GetValuation(w http.ResponseWriter, r *http.Request) {
	// URLパスから企業IDを抽出
	path := r.URL.Path
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) < 5 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	targetCompanyID := parts[len(parts)-1]

	if targetCompanyID == "" {
		http.Error(w, "company_id is required", http.StatusBadRequest)
		return
	}

	// 企業情報を取得するため、実装が必要
	// ここでは簡易的にエラーを返す
	http.Error(w, "valuation endpoint not fully implemented", http.StatusNotImplemented)
}

// getCompanyIDFromContext コンテキストから企業IDを取得する
func getCompanyIDFromContext(ctx context.Context) string {
	if companyID, ok := ctx.Value(companyIDKey).(string); ok {
		return companyID
	}
	return ""
}
