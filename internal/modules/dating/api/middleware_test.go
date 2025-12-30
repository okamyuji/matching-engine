package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestAuthMiddleware_NoAuth(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// 認証ヘッダーがない場合は401を返す
	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// コンテキストからユーザーIDを取得して確認
		userID := getUserIDFromContext(r.Context())
		if userID != "test-user" {
			t.Errorf("userID = %v, want test-user", userID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	// 有効なJWTトークンを生成
	token := generateTestJWT("test-user", time.Now().Add(1*time.Hour).Unix())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("AuthMiddleware() status = %v, want %v", w.Code, http.StatusOK)
	}
}

// generateTestJWT テスト用の有効なJWTトークンを生成する
func generateTestJWT(userID string, exp int64) string {
	// ヘッダー
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		panic(err) // テスト用なのでパニックでOK
	}
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// ペイロード
	payload := map[string]any{
		"user_id": userID,
		"exp":     exp,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		panic(err) // テスト用なのでパニックでOK
	}
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// 署名
	message := headerB64 + "." + payloadB64
	mac := hmac.New(sha256.New, jwtSecret)
	mac.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return message + "." + signature
}

func TestLoggingMiddleware(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := LoggingMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("LoggingMiddleware() status = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	rw.WriteHeader(http.StatusNotFound)

	if rw.statusCode != http.StatusNotFound {
		t.Errorf("responseWriter.statusCode = %v, want %v", rw.statusCode, http.StatusNotFound)
	}
}

func TestResponseWriter_Write(t *testing.T) {
	w := httptest.NewRecorder()
	rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

	if rw.statusCode != http.StatusOK {
		t.Errorf("Initial statusCode = %v, want %v", rw.statusCode, http.StatusOK)
	}

	data := []byte("test response")
	n, err := rw.Write(data)

	if err != nil {
		t.Errorf("responseWriter.Write() error = %v", err)
	}
	if n != len(data) {
		t.Errorf("responseWriter.Write() wrote %v bytes, want %v", n, len(data))
	}
}

func TestLoggingMiddleware_DifferentStatusCodes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"status 200", http.StatusOK},
		{"status 400", http.StatusBadRequest},
		{"status 401", http.StatusUnauthorized},
		{"status 500", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			middleware := LoggingMiddleware(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("LoggingMiddleware() status = %v, want %v", w.Code, tt.statusCode)
			}
		})
	}
}

func TestJWTError_Error(t *testing.T) {
	err := &jwtError{message: "invalid token"}

	if err.Error() != "invalid token" {
		t.Errorf("jwtError.Error() = %v, want 'invalid token'", err.Error())
	}
}

func TestAuthMiddleware_InvalidBearerFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(handler)

	// 過去の有効期限でトークンを生成
	token := generateTestJWT("test-user", time.Now().Add(-1*time.Hour).Unix())

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("AuthMiddleware() status = %v, want %v for expired token", w.Code, http.StatusUnauthorized)
	}
}
