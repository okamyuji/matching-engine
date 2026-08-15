// Package auth HS256 の JWT を最小限に扱う。外部ライブラリを使わず、
// dating と ma の両モジュールが同じ検証規則を共有するために置く。
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// ErrInvalidToken トークンの形式・署名・有効期限が不正
var ErrInvalidToken = errors.New("invalid token")

// defaultSecret 開発環境用の既定秘密鍵。本番では JWT_SECRET を必ず設定する
const defaultSecret = "development-secret-key-change-in-production"

// SecretFromEnv JWT_SECRET 環境変数から秘密鍵を返す。未設定なら開発用の既定値を返す
func SecretFromEnv() []byte {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return []byte(s)
	}
	return []byte(defaultSecret)
}

// Sign クレームを HS256 で署名した JWT を返す
func Sign(secret []byte, claims map[string]any) (string, error) {
	headerJSON, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal claims: %w", err)
	}
	message := base64.RawURLEncoding.EncodeToString(headerJSON) + "." + base64.RawURLEncoding.EncodeToString(payloadJSON)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(message))
	return message + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil)), nil
}

// Verify 署名と有効期限（exp、Unix秒）を検証し、クレームを返す
func Verify(secret []byte, token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("%w: invalid token format", ErrInvalidToken)
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid signature encoding", ErrInvalidToken)
	}
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return nil, fmt.Errorf("%w: signature verification failed", ErrInvalidToken)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("%w: invalid payload encoding", ErrInvalidToken)
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("%w: invalid claims format", ErrInvalidToken)
	}
	if exp, ok := claims["exp"]; ok {
		expF, isNum := exp.(float64)
		if !isNum {
			return nil, fmt.Errorf("%w: invalid exp", ErrInvalidToken)
		}
		if expF > 0 && time.Now().Unix() > int64(expF) {
			return nil, fmt.Errorf("%w: token expired", ErrInvalidToken)
		}
	}
	return claims, nil
}

// StringClaim クレームから文字列値を取り出す。無い、または文字列でなければ空文字を返す
func StringClaim(claims map[string]any, key string) string {
	v, ok := claims[key].(string)
	if !ok {
		return ""
	}
	return v
}
