package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger データベースの疎通確認だけを行う最小インターフェース
type Pinger interface {
	Ping(ctx context.Context) error
}

// Handler ヘルスチェックハンドラー
type Handler struct {
	db Pinger
}

// NewHandler 新しいヘルスチェックハンドラーを作成する
func NewHandler(db Pinger) *Handler {
	return &Handler{db: db}
}

// LivenessHandler アプリケーションが起動しているかチェックする
// GET /health/live
func (h *Handler) LivenessHandler(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ReadinessHandler アプリケーションがリクエストを受け付けられるかチェックする
// GET /health/ready
func (h *Handler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"error":  "database connection failed",
		})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func writeJSON(w http.ResponseWriter, status int, body map[string]string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
