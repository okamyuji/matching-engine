package api

import "net/http"

// SetupRoutes Dating APIのHTTPルートを設定する。全ルートを認証ミドルウェアで包む
func SetupRoutes(mux *http.ServeMux, h *Handler) {
	mux.Handle("GET /api/v1/dating/matches", AuthMiddleware(http.HandlerFunc(h.GetMatches)))
	mux.Handle("POST /api/v1/dating/likes", AuthMiddleware(http.HandlerFunc(h.SendLike)))
	mux.Handle("GET /api/v1/dating/likes/received", AuthMiddleware(http.HandlerFunc(h.GetReceivedLikes)))
	mux.Handle("GET /api/v1/dating/matches/mutual", AuthMiddleware(http.HandlerFunc(h.GetMutualMatches)))
}
