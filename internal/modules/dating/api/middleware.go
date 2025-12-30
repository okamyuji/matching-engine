package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// JWT秘密鍵（環境変数から取得、デフォルトは開発用）
var jwtSecret = getJWTSecret()

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		// 開発環境用のデフォルト秘密鍵
		secret = "development-secret-key-change-in-production"
	}
	return []byte(secret)
}

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
	// トークンを3つの部分に分割（header.payload.signature）
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, &jwtError{"invalid token format"}
	}

	// 署名を検証
	message := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, &jwtError{"invalid signature encoding"}
	}

	// HMAC SHA256で署名を計算
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(message))
	expectedSignature := mac.Sum(nil)

	// 署名を比較
	if !hmac.Equal(signature, expectedSignature) {
		return nil, &jwtError{"signature verification failed"}
	}

	// ペイロードをデコード
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, &jwtError{"invalid payload encoding"}
	}

	// クレームをパース
	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, &jwtError{"invalid claims format"}
	}

	// 有効期限を確認
	if claims.Exp > 0 && time.Now().Unix() > claims.Exp {
		return nil, &jwtError{"token expired"}
	}

	// ユーザーIDが存在することを確認
	if claims.UserID == "" {
		return nil, &jwtError{"missing user_id in claims"}
	}

	return &claims, nil
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
