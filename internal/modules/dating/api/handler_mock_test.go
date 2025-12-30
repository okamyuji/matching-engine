package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourorg/matching-engine/internal/modules/dating/application"
	"github.com/yourorg/matching-engine/internal/modules/dating/infrastructure/repository"
)

// mockMatchingService モックマッチングサービス
type mockMatchingService struct {
	results       []*application.MatchResult
	mutualResults []*application.MutualMatchResult
	err           error
}

func (m *mockMatchingService) FindMatches(ctx context.Context, userID string, limit int) ([]*application.MatchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.results, nil
}

func (m *mockMatchingService) GetMutualMatches(ctx context.Context, userID string) ([]*application.MutualMatchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.mutualResults, nil
}

// mockLikeService モックいいねサービス
type mockLikeService struct {
	likeResp *application.LikeResponse
	likes    []*repository.Like
	err      error
}

func (m *mockLikeService) SendLike(ctx context.Context, fromUserID, toUserID string) (*application.LikeResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.likeResp, nil
}

func (m *mockLikeService) GetReceivedLikes(ctx context.Context, userID string) ([]*repository.Like, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.likes, nil
}

// TestGetMatches_Success モックを使用した正常系テスト
func TestGetMatches_Success(t *testing.T) {
	mockService := &mockMatchingService{
		results: []*application.MatchResult{
			{
				UserID:    "user2",
				Score:     0.95,
				Rank:      1,
				Breakdown: map[string]float64{"age": 0.95},
			},
		},
	}

	handler := NewHandler(mockService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/matches?limit=10", nil)
	// handler.goのuserIDKeyを使用してコンテキストに値を設定
	ctx := context.WithValue(req.Context(), userIDKey, "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetMatches(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetMatches() status = %v, want %v", w.Code, http.StatusOK)
	}

	var results []*application.MatchResult
	if err := json.NewDecoder(w.Body).Decode(&results); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("GetMatches() returned %d results, want 1", len(results))
	}

	if results[0].UserID != "user2" {
		t.Errorf("GetMatches() UserID = %v, want user2", results[0].UserID)
	}
}

func TestSendLike_Success_NoMatch(t *testing.T) {
	mockService := &mockLikeService{
		likeResp: &application.LikeResponse{
			Matched: false,
		},
	}

	handler := NewHandler(nil, mockService)

	reqBody := application.LikeRequest{
		TargetUserID: "user2",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dating/likes", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), userIDKey, "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.SendLike(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SendLike() status = %v, want %v", w.Code, http.StatusOK)
	}

	var resp application.LikeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Matched {
		t.Errorf("SendLike() Matched = true, want false")
	}
}

func TestSendLike_Success_Matched(t *testing.T) {
	mockService := &mockLikeService{
		likeResp: &application.LikeResponse{
			Matched: true,
			MatchID: "match_123",
		},
	}

	handler := NewHandler(nil, mockService)

	reqBody := application.LikeRequest{
		TargetUserID: "user2",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dating/likes", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), userIDKey, "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.SendLike(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("SendLike() status = %v, want %v", w.Code, http.StatusOK)
	}

	var resp application.LikeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Matched {
		t.Errorf("SendLike() Matched = false, want true")
	}

	if resp.MatchID != "match_123" {
		t.Errorf("SendLike() MatchID = %v, want match_123", resp.MatchID)
	}
}

func TestGetReceivedLikes_Success(t *testing.T) {
	mockService := &mockLikeService{
		likes: []*repository.Like{
			{
				ID:         "like_1",
				FromUserID: "user2",
				ToUserID:   "user1",
			},
		},
	}

	handler := NewHandler(nil, mockService)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/likes/received", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetReceivedLikes(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetReceivedLikes() status = %v, want %v", w.Code, http.StatusOK)
	}

	var likes []*repository.Like
	if err := json.NewDecoder(w.Body).Decode(&likes); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(likes) != 1 {
		t.Errorf("GetReceivedLikes() returned %d likes, want 1", len(likes))
	}

	if likes[0].FromUserID != "user2" {
		t.Errorf("GetReceivedLikes() FromUserID = %v, want user2", likes[0].FromUserID)
	}
}

func TestGetMutualMatches_Success(t *testing.T) {
	mockService := &mockMatchingService{
		mutualResults: []*application.MutualMatchResult{
			{
				MatchID:   "match_123",
				UserIDA:   "user1",
				UserIDB:   "user2",
				Score:     0.85,
				Breakdown: map[string]float64{"age": 0.85},
				MatchedAt: "2024-01-01T00:00:00Z",
			},
		},
	}

	handler := NewHandler(mockService, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/matches/mutual", nil)
	ctx := context.WithValue(req.Context(), userIDKey, "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.GetMutualMatches(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GetMutualMatches() status = %v, want %v", w.Code, http.StatusOK)
	}

	var matches []*application.MutualMatchResult
	if err := json.NewDecoder(w.Body).Decode(&matches); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("GetMutualMatches() returned %d matches, want 1", len(matches))
	}

	if len(matches) > 0 && matches[0].MatchID != "match_123" {
		t.Errorf("GetMutualMatches() MatchID = %v, want match_123", matches[0].MatchID)
	}
}

// TestHandler_ErrorCases エラーケースのテスト
func TestSendLike_EmptyTargetUserID(t *testing.T) {
	handler := NewHandler(nil, nil)

	reqBody := application.LikeRequest{
		TargetUserID: "",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dating/likes", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), userIDKey, "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.SendLike(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("SendLike() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestSendLike_InvalidJSON_WithAuth(t *testing.T) {
	handler := NewHandler(nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dating/likes", bytes.NewReader([]byte("invalid json")))
	ctx := context.WithValue(req.Context(), userIDKey, "user1")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.SendLike(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("SendLike() status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}
