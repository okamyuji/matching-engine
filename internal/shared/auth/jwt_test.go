package auth

import (
	"errors"
	"testing"
	"time"
)

func TestSignAndVerify_RoundTrip(t *testing.T) {
	secret := []byte("s3cret")
	token, err := Sign(secret, map[string]any{"user_id": "u1", "exp": time.Now().Add(time.Hour).Unix()})
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	claims, err := Verify(secret, token)
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if got := StringClaim(claims, "user_id"); got != "u1" {
		t.Errorf("user_id = %q, want u1", got)
	}
	if got := StringClaim(claims, "missing"); got != "" {
		t.Errorf("missing claim = %q, want empty", got)
	}
}

func TestVerify_Rejects(t *testing.T) {
	secret := []byte("s3cret")
	valid, _ := Sign(secret, map[string]any{"user_id": "u1"})
	expired, _ := Sign(secret, map[string]any{"user_id": "u1", "exp": time.Now().Add(-time.Minute).Unix()})
	badExp, _ := Sign(secret, map[string]any{"user_id": "u1", "exp": "soon"})
	otherSecret, _ := Sign([]byte("other"), map[string]any{"user_id": "u1"})

	cases := map[string]string{
		"format":     "a.b",
		"signature":  valid[:len(valid)-2] + "xx",
		"sigEncode":  "a.b.%%%",
		"payload":    "a.%%%." + valid[len(valid)-10:],
		"expired":    expired,
		"badExp":     badExp,
		"wrongKey":   otherSecret,
		"notJSONPay": "eyJhbGciOiJIUzI1NiJ9.bm90anNvbg.sig",
	}
	for name, tok := range cases {
		if _, err := Verify(secret, tok); !errors.Is(err, ErrInvalidToken) {
			t.Errorf("%s: err = %v, want ErrInvalidToken", name, err)
		}
	}
}

func TestSecretFromEnv(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	if string(SecretFromEnv()) != defaultSecret {
		t.Error("既定の秘密鍵が返るべき")
	}
	t.Setenv("JWT_SECRET", "abc")
	if string(SecretFromEnv()) != "abc" {
		t.Error("環境変数の秘密鍵が返るべき")
	}
}

func TestVerify_ExpZeroMeansNoExpiry(t *testing.T) {
	secret := []byte("s3cret")
	tok, _ := Sign(secret, map[string]any{"user_id": "u1", "exp": 0})
	if _, err := Verify(secret, tok); err != nil {
		t.Errorf("exp=0 は有効期限なしとして扱うべき: %v", err)
	}
	future, _ := Sign(secret, map[string]any{"user_id": "u1", "exp": time.Now().Add(2 * time.Second).Unix()})
	if _, err := Verify(secret, future); err != nil {
		t.Errorf("未来の exp は有効: %v", err)
	}
}
