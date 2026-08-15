package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"github.com/okamyuji/matching-engine/internal/shared/auth"
)

// jwtSecret JWT秘密鍵（環境変数 JWT_SECRET、未設定なら開発用の既定値）
var jwtSecret = auth.SecretFromEnv()

// AuthMiddleware 認証用ミドルウェア
// Authorization: Bearer <token> 形式のJWTを検証し、company_id クレームをコンテキストへ入れる
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}
		claims, err := auth.Verify(jwtSecret, parts[1])
		if err != nil {
			slog.Warn("JWT validation failed", "error", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		companyID := auth.StringClaim(claims, "company_id")
		if companyID == "" {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), companyIDKey, companyID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
