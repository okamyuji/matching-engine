package app

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/okamyuji/matching-engine/internal/testutil"
)

func TestNewRouter(t *testing.T) {
	td := testutil.GetSharedTestDB(t)
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	dating := filepath.Join(root, "configs", "dating", "matching.json")
	ma := filepath.Join(root, "configs", "ma", "matching.json")

	h, err := NewRouter(td.Pool, Options{DatingConfigPath: dating, MAConfigPath: ma})
	if err != nil {
		t.Fatalf("NewRouter: %v", err)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health/live", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("health status = %d", rec.Code)
	}
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/dating/matches", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("認証なしの API は 401 のはず: %d", rec.Code)
	}

	if _, err := NewRouter(td.Pool, Options{DatingConfigPath: "/nonexistent.json", MAConfigPath: ma}); err == nil {
		t.Error("dating 設定が無ければエラーになるべき")
	}
	if _, err := NewRouter(td.Pool, Options{DatingConfigPath: dating, MAConfigPath: "/nonexistent.json"}); err == nil {
		t.Error("ma 設定が無ければエラーになるべき")
	}
}
