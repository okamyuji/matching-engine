package api

import "net/http"

// SetupRoutes Dating APIのHTTPルートを設定する
func SetupRoutes(mux *http.ServeMux, h *Handler) {
	// 全ルートをミドルウェアでラップ
	mux.HandleFunc("GET /api/v1/dating/matches", h.GetMatches)
	mux.HandleFunc("POST /api/v1/dating/likes", h.SendLike)
	mux.HandleFunc("GET /api/v1/dating/likes/received", h.GetReceivedLikes)
	mux.HandleFunc("GET /api/v1/dating/matches/mutual", h.GetMutualMatches)
}
