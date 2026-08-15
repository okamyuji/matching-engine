package api

import (
	"net/http"
)

// SetupRoutes M&A APIのルートを設定する。全ルートを認証ミドルウェアで包む
func SetupRoutes(mux *http.ServeMux, h *Handler) {
	mux.Handle("GET /api/v1/ma/targets", AuthMiddleware(http.HandlerFunc(h.GetTargets)))
	mux.Handle("POST /api/v1/ma/interests", AuthMiddleware(http.HandlerFunc(h.SendInterest)))
	mux.Handle("GET /api/v1/ma/interests/received", AuthMiddleware(http.HandlerFunc(h.GetReceivedInterests)))
	mux.Handle("GET /api/v1/ma/matches", AuthMiddleware(http.HandlerFunc(h.GetMatches)))
	mux.Handle("GET /api/v1/ma/valuation/{id}", AuthMiddleware(http.HandlerFunc(h.GetValuation)))
}
