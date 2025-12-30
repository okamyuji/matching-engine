package api

import (
	"net/http"
)

// SetupRoutes M&A APIのルートを設定する
func SetupRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("GET /api/v1/ma/targets", h.GetTargets)
	mux.HandleFunc("POST /api/v1/ma/interests", h.SendInterest)
	mux.HandleFunc("GET /api/v1/ma/interests/received", h.GetReceivedInterests)
	mux.HandleFunc("GET /api/v1/ma/matches", h.GetMatches)
	mux.HandleFunc("GET /api/v1/ma/valuation/{id}", h.GetValuation)
}
