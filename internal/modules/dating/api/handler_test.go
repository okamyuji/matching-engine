package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okamyuji/matching-engine/internal/modules/dating/application"
)

func TestNewHandler(t *testing.T) {
	handler := NewHandler(nil, nil)
	if handler == nil {
		t.Error("NewHandler() returned nil")
	}
}

func TestGetMatches_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/matches", nil)
	w := httptest.NewRecorder()

	handler.GetMatches(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetMatches() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestSendLike_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dating/likes", nil)
	w := httptest.NewRecorder()

	handler.SendLike(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("SendLike() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetReceivedLikes_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/likes/received", nil)
	w := httptest.NewRecorder()

	handler.GetReceivedLikes(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetReceivedLikes() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetMutualMatches_Unauthorized(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/matches/mutual", nil)
	w := httptest.NewRecorder()

	handler.GetMutualMatches(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("GetMutualMatches() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	userID := getUserIDFromContext(nil)
	if userID != "" {
		t.Errorf("getUserIDFromContext() = %v, want empty string", userID)
	}
}

func TestGetMatches_QueryLimit(t *testing.T) {
	handler := NewHandler(nil, nil)

	tests := []struct {
		name  string
		query string
	}{
		{"with valid limit", "?limit=50"},
		{"with invalid limit", "?limit=abc"},
		{"with zero limit", "?limit=0"},
		{"with negative limit", "?limit=-5"},
		{"no limit", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/matches"+tt.query, nil)
			w := httptest.NewRecorder()

			handler.GetMatches(w, req)

			if w.Code != http.StatusUnauthorized {
				t.Errorf("GetMatches() status = %v, want %v", w.Code, http.StatusUnauthorized)
			}
		})
	}
}

func TestSendLike_InvalidJSON(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dating/likes", nil)
	w := httptest.NewRecorder()

	handler.SendLike(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("SendLike() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestSendLike_WithInvalidBody(t *testing.T) {
	handler := NewHandler(nil, nil)

	// Invalid JSON body
	body := bytes.NewReader([]byte("{invalid json"))
	req := httptest.NewRequest(http.MethodPost, "/api/v1/dating/likes", body)
	w := httptest.NewRecorder()

	handler.SendLike(w, req)

	// Should return unauthorized since getUserIDFromContext returns empty
	if w.Code != http.StatusUnauthorized {
		t.Errorf("SendLike() status = %v, want %v", w.Code, http.StatusUnauthorized)
	}
}

func TestLikeRequest_JSONMarshal(t *testing.T) {
	req := application.LikeRequest{
		TargetUserID: "user123",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded application.LikeRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.TargetUserID != "user123" {
		t.Errorf("TargetUserID = %v, want user123", decoded.TargetUserID)
	}
}

func TestMatchResult_JSONMarshal(t *testing.T) {
	result := application.MatchResult{
		UserID: "user123",
		Score:  0.95,
		Rank:   1,
		Breakdown: map[string]float64{
			"test": 0.95,
		},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	if len(data) == 0 {
		t.Error("json.Marshal() returned empty data")
	}
}

func TestLikeResponse_JSONMarshal(t *testing.T) {
	tests := []struct {
		name     string
		response application.LikeResponse
	}{
		{
			name: "matched",
			response: application.LikeResponse{
				Matched: true,
				MatchID: "match_123",
			},
		},
		{
			name: "not matched",
			response: application.LikeResponse{
				Matched: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}
			if len(data) == 0 {
				t.Error("json.Marshal() returned empty data")
			}

			var decoded application.LikeResponse
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}
			if decoded.Matched != tt.response.Matched {
				t.Errorf("Matched = %v, want %v", decoded.Matched, tt.response.Matched)
			}
		})
	}
}
