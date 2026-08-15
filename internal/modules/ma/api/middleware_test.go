package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/shared/auth"
)

func TestAuthMiddleware(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := getCompanyIDFromContext(r.Context())
		if id == "" {
			t.Error("company_id がコンテキストに無い")
		}
		w.WriteHeader(http.StatusNoContent)
	})
	handler := AuthMiddleware(next)

	okToken, err := auth.Sign(jwtSecret, map[string]any{"company_id": "c1", "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	userToken, _ := auth.Sign(jwtSecret, map[string]any{"user_id": "u1"})

	cases := []struct {
		name   string
		header string
		want   int
	}{
		{"ヘッダーなし", "", http.StatusUnauthorized},
		{"形式不正", "Token abc", http.StatusUnauthorized},
		{"署名不正", "Bearer a.b.c", http.StatusUnauthorized},
		{"company_id 欠落", "Bearer " + userToken, http.StatusUnauthorized},
		{"正常", "Bearer " + okToken, http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tc.want {
				t.Errorf("status = %d, want %d", rec.Code, tc.want)
			}
		})
	}
}
