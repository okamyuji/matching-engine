package health

import (
	"encoding/json"
	"net/http"

	"github.com/uptrace/bun"
)

// Handler ヘルスチェックハンドラー
type Handler struct {
	db *bun.DB
}

// NewHandler 新しいヘルスチェックハンドラーを作成する
func NewHandler(db *bun.DB) *Handler {
	return &Handler{db: db}
}

// LivenessHandler アプリケーションが起動しているかチェックする
// GET /health/live
func (h *Handler) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ReadinessHandler アプリケーションがリクエストを受け付けられるかチェックする
// GET /health/ready
func (h *Handler) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// データベース接続をチェック
	if err := h.db.Ping(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)

		response := map[string]string{
			"status": "unavailable",
			"error":  "database connection failed",
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status": "ready",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
