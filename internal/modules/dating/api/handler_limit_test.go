package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/okamyuji/matching-engine/internal/modules/dating/application"
)

// recordingMatchingService FindMatches に渡された limit を記録する
type recordingMatchingService struct {
	mockMatchingService
	gotLimit int
}

func (r *recordingMatchingService) FindMatches(ctx context.Context, userID string, limit int) ([]*application.MatchResult, error) {
	r.gotLimit = limit
	return r.mockMatchingService.FindMatches(ctx, userID, limit)
}

func TestGetMatches_LimitParsing(t *testing.T) {
	cases := []struct {
		query string
		want  int
	}{
		{"", 20},
		{"?limit=50", 50},
		{"?limit=1", 1},
		{"?limit=0", 20},
		{"?limit=-5", 20},
		{"?limit=abc", 20},
	}
	for _, tc := range cases {
		t.Run(tc.query, func(t *testing.T) {
			svc := &recordingMatchingService{}
			h := NewHandler(svc, &mockLikeService{})
			req := httptest.NewRequest(http.MethodGet, "/api/v1/dating/matches"+tc.query, nil)
			req = req.WithContext(context.WithValue(req.Context(), userIDKey, "user1"))
			w := httptest.NewRecorder()
			h.GetMatches(w, req)
			if w.Code != http.StatusOK {
				t.Fatalf("status = %d", w.Code)
			}
			if svc.gotLimit != tc.want {
				t.Errorf("limit = %d, want %d", svc.gotLimit, tc.want)
			}
		})
	}
}
