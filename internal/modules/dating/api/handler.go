package api

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yourorg/matching-engine/internal/modules/dating/application"
)

// Handler Dating API のHTTPリクエストを処理する
type Handler struct {
	matchingService application.MatchingServiceInterface
	likeService     application.LikeServiceInterface
}

// NewHandler 新しいHandlerを作成する
func NewHandler(
	matchingService application.MatchingServiceInterface,
	likeService application.LikeServiceInterface,
) *Handler {
	return &Handler{
		matchingService: matchingService,
		likeService:     likeService,
	}
}

// GetMatches ユーザーのマッチング候補を取得する
// GET /api/v1/dating/matches
func (h *Handler) GetMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)

	if userID == "" {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	matches, err := h.matchingService.FindMatches(ctx, userID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matches); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// SendLike 別のユーザーにいいねを送る
// POST /api/v1/dating/likes
func (h *Handler) SendLike(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)

	if userID == "" {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	var req application.LikeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.TargetUserID == "" {
		http.Error(w, "target_user_id is required", http.StatusBadRequest)
		return
	}

	resp, err := h.likeService.SendLike(ctx, userID, req.TargetUserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetReceivedLikes ユーザーが受け取った全てのいいねを取得する
// GET /api/v1/dating/likes/received
func (h *Handler) GetReceivedLikes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)

	if userID == "" {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	likes, err := h.likeService.GetReceivedLikes(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(likes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetMutualMatches ユーザーの相互マッチを取得する
// GET /api/v1/dating/matches/mutual
func (h *Handler) GetMutualMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)

	if userID == "" {
		http.Error(w, "user not authenticated", http.StatusUnauthorized)
		return
	}

	matches, err := h.matchingService.GetMutualMatches(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(matches); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// userIDContextKey コンテキストからユーザーIDを取得するためのキー
type userIDContextKey string

const userIDKey userIDContextKey = "user_id"

// getUserIDFromContext リクエストコンテキストからユーザーIDを取得する
// AuthMiddlewareによってコンテキストに追加されたユーザーIDを取得する
func getUserIDFromContext(ctx any) string {
	if ctx == nil {
		return ""
	}

	if reqCtx, ok := ctx.(context.Context); ok {
		if userID, ok := reqCtx.Value(userIDKey).(string); ok {
			return userID
		}
	}

	return ""
}
