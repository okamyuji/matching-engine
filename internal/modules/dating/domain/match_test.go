package domain

import (
	"testing"
	"time"
)

func TestMatch_Creation(t *testing.T) {
	tests := []struct {
		name  string
		match Match
	}{
		{
			name: "valid match with breakdown",
			match: Match{
				ID:      "match123",
				UserIDA: "user1",
				UserIDB: "user2",
				Score:   0.85,
				Breakdown: map[string]float64{
					"age_similarity":         0.90,
					"location_match":         1.00,
					"hobby_similarity":       0.75,
					"marriage_compatibility": 0.85,
					"height_preference":      0.80,
					"smoking_compatibility":  1.00,
					"drinking_compatibility": 0.90,
				},
				MatchedAt: time.Now(),
			},
		},
		{
			name: "minimal match",
			match: Match{
				ID:        "match456",
				UserIDA:   "user3",
				UserIDB:   "user4",
				Score:     0.50,
				Breakdown: map[string]float64{},
				MatchedAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.match

			if m.ID == "" {
				t.Error("ID should not be empty")
			}
			if m.UserIDA == "" {
				t.Error("UserIDA should not be empty")
			}
			if m.UserIDB == "" {
				t.Error("UserIDB should not be empty")
			}
			if m.Score < 0 || m.Score > 1 {
				t.Errorf("Score %v should be between 0 and 1", m.Score)
			}
			if m.MatchedAt.IsZero() {
				t.Error("MatchedAt should not be zero")
			}

			if tt.name == "valid match with breakdown" {
				if len(m.Breakdown) != 7 {
					t.Errorf("Breakdown length = %v, want 7", len(m.Breakdown))
				}
			}
		})
	}
}
