package application

import "testing"

func TestMatchResult_Creation(t *testing.T) {
	result := &MatchResult{
		UserID: "user123",
		Score:  0.85,
		Rank:   1,
		Breakdown: map[string]float64{
			"age_similarity":   0.90,
			"hobby_similarity": 0.80,
		},
	}

	if result.UserID != "user123" {
		t.Errorf("UserID = %v, want user123", result.UserID)
	}
	if result.Score != 0.85 {
		t.Errorf("Score = %v, want 0.85", result.Score)
	}
	if result.Rank != 1 {
		t.Errorf("Rank = %v, want 1", result.Rank)
	}
	if len(result.Breakdown) != 2 {
		t.Errorf("Breakdown length = %v, want 2", len(result.Breakdown))
	}
}

func TestLikeRequest_Creation(t *testing.T) {
	req := &LikeRequest{
		TargetUserID: "user456",
	}

	if req.TargetUserID != "user456" {
		t.Errorf("TargetUserID = %v, want user456", req.TargetUserID)
	}
}

func TestLikeResponse_Creation(t *testing.T) {
	tests := []struct {
		name     string
		response LikeResponse
	}{
		{
			name: "matched response",
			response: LikeResponse{
				Matched: true,
				MatchID: "match123",
			},
		},
		{
			name: "not matched response",
			response: LikeResponse{
				Matched: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "matched response" {
				if !tt.response.Matched {
					t.Error("Matched should be true")
				}
				if tt.response.MatchID != "match123" {
					t.Errorf("MatchID = %v, want match123", tt.response.MatchID)
				}
			} else {
				if tt.response.Matched {
					t.Error("Matched should be false")
				}
				if tt.response.MatchID != "" {
					t.Error("MatchID should be empty for non-matched response")
				}
			}
		})
	}
}

func TestLikeResponse_MatchedWithID(t *testing.T) {
	resp := LikeResponse{
		Matched: true,
		MatchID: "match_123",
	}

	if !resp.Matched {
		t.Error("Matched should be true")
	}
	if resp.MatchID == "" {
		t.Error("MatchID should not be empty")
	}
}

func TestLikeResponse_NotMatched(t *testing.T) {
	resp := LikeResponse{
		Matched: false,
	}

	if resp.Matched {
		t.Error("Matched should be false")
	}
	if resp.MatchID != "" {
		t.Error("MatchID should be empty when not matched")
	}
}

func TestMatchResult_EmptyBreakdown(t *testing.T) {
	result := &MatchResult{
		UserID:    "user1",
		Score:     0.5,
		Rank:      5,
		Breakdown: make(map[string]float64),
	}

	if result.UserID != "user1" {
		t.Errorf("UserID = %v, want user1", result.UserID)
	}
	if result.Score != 0.5 {
		t.Errorf("Score = %v, want 0.5", result.Score)
	}
	if result.Rank != 5 {
		t.Errorf("Rank = %v, want 5", result.Rank)
	}
	if len(result.Breakdown) != 0 {
		t.Errorf("Breakdown length = %v, want 0", len(result.Breakdown))
	}
}
