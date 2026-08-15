package api

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/okamyuji/matching-engine/internal/shared/auth"
)

// jwtSecret JWT秘密鍵（環境変数 JWT_SECRET、未設定なら開発用の既定値）
var jwtSecret = auth.SecretFromEnv()

// jwtClaims JWT クレーム構造体
type jwtClaims struct {
	UserID string `json:"user_id"`
	Exp    int64  `json:"exp"` // Unix timestamp
}

// AuthMiddleware 認証用ミドルウェア
// Authorization: Bearer <token> 形式のJWTトークンを検証する
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Authorizationヘッダーから取得
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// 2. Bearer トークンを抽出
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		// 3. JWTトークンを検証
		claims, err := validateJWT(token)
		if err != nil {
			slog.Warn("JWT validation failed", "error", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// 4. コンテキストにユーザーIDを追加
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// validateJWT JWTトークンを検証してクレームを返す
func validateJWT(token string) (*jwtClaims, error) {
	claims, err := auth.Verify(jwtSecret, token)
	if err != nil {
		return nil, &jwtError{strings.TrimPrefix(err.Error(), auth.ErrInvalidToken.Error()+": ")}
	}
	userID := auth.StringClaim(claims, "user_id")
	if userID == "" {
		return nil, &jwtError{"missing user_id in claims"}
	}
	var exp int64
	if v, ok := claims["exp"].(float64); ok {
		exp = int64(v)
	}
	return &jwtClaims{UserID: userID, Exp: exp}, nil
}

// jwtError JWT検証エラー
type jwtError struct {
	message string
}

func (e *jwtError) Error() string {
	return e.message
}

// LoggingMiddleware HTTPリクエストをログ出力する
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// ステータスコードをキャプチャするためのレスポンスライター作成
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 次のハンドラを呼び出し
		next.ServeHTTP(rw, r)

		// リクエストをログ出力
		duration := time.Since(start)
		slog.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", duration.Milliseconds(),
		)
	})
}

// responseWriter ステータスコードをキャプチャするためのhttp.ResponseWriterラッパー
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader ステータスコードをキャプチャする
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
