package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/matching-engine/internal/testutil"
)

func TestLivenessHandler(t *testing.T) {
	// testutilのSharedTestDBを使用
	td := testutil.GetSharedTestDB(t)

	handler := NewHandler(td.DB)

	req := httptest.NewRequest(http.MethodGet, "/health/live", nil)
	w := httptest.NewRecorder()

	handler.LivenessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("status = %s, want ok", response["status"])
	}
}

func TestReadinessHandler_Success(t *testing.T) {
	// testutilのSharedTestDBを使用
	td := testutil.GetSharedTestDB(t)

	handler := NewHandler(td.DB)

	req := httptest.NewRequest(http.MethodGet, "/health/ready", nil)
	w := httptest.NewRecorder()

	handler.ReadinessHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["status"] != "ready" {
		t.Errorf("status = %s, want ready", response["status"])
	}
}
