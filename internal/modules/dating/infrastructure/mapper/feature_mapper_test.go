package mapper

import (
	"math"
	"testing"
	"time"

	"github.com/okamyuji/matching-engine/internal/modules/dating/domain"
)

func TestDatingFeatureMapper_ToFeatureVector(t *testing.T) {
	mapper := NewDatingFeatureMapper()

	now := time.Now()
	user := &domain.User{
		ID:           "user123",
		Nickname:     "Alice",
		Gender:       domain.GenderFemale,
		BirthDate:    now.AddDate(-25, 0, 0), // 25 years old
		Prefecture:   domain.PrefectureTokyo,
		Verified:     true,
		EloRating:    1500,
		CreatedAt:    now.AddDate(0, -6, 0), // 6 months ago
		LastActiveAt: now.AddDate(0, 0, -7), // 7 days ago
	}

	profile := &domain.Profile{
		UserID:         "user123",
		Height:         165,
		BodyType:       domain.BodyTypeAverage,
		Education:      domain.EducationUniversity,
		Occupation:     "Engineer",
		IncomeLevel:    500, // 500万円
		MarriageDesire: domain.MarriageWantEventually,
		ChildrenDesire: domain.ChildrenWant,
		Smoking:        domain.SmokingNonSmoker,
		Drinking:       domain.DrinkingSocial,
		Tags: []domain.ProfileTag{
			{UserID: "user123", Tag: "sports"},
			{UserID: "user123", Tag: "travel"},
			{UserID: "user123", Tag: "music"},
		},
		SelfIntroduction: "Hello!",
	}

	fv := mapper.ToFeatureVector(user, profile)

	// Check basic properties
	if fv.ID != "user123" {
		t.Errorf("ID = %v, want user123", fv.ID)
	}
	if fv.Type != "dating_user" {
		t.Errorf("Type = %v, want dating_user", fv.Type)
	}

	// Check numerical features
	if _, ok := fv.Numerical["age"]; !ok {
		t.Error("age feature not found")
	}
	if _, ok := fv.Numerical["height"]; !ok {
		t.Error("height feature not found")
	}
	if _, ok := fv.Numerical["income"]; !ok {
		t.Error("income feature not found")
	}
	if _, ok := fv.Numerical["elo"]; !ok {
		t.Error("elo feature not found")
	}
	if _, ok := fv.Numerical["activity"]; !ok {
		t.Error("activity feature not found")
	}
	if _, ok := fv.Numerical["recency"]; !ok {
		t.Error("recency feature not found")
	}

	// Check categorical features
	if fv.Categorical["prefecture"]["tokyo"] != 1.0 {
		t.Error("prefecture categorical feature incorrect")
	}
	if fv.Categorical["body_type"]["average"] != 1.0 {
		t.Error("body_type categorical feature incorrect")
	}
	if fv.Categorical["education"]["university"] != 1.0 {
		t.Error("education categorical feature incorrect")
	}
	if fv.Categorical["marriage_desire"]["want_eventually"] != 1.0 {
		t.Error("marriage_desire categorical feature incorrect")
	}
	if fv.Categorical["smoking"]["non_smoker"] != 1.0 {
		t.Error("smoking categorical feature incorrect")
	}
	if fv.Categorical["drinking"]["social"] != 1.0 {
		t.Error("drinking categorical feature incorrect")
	}

	// Check sparse features (tags)
	if fv.Sparse["tags"]["sports"] != 1.0 {
		t.Error("sports tag not found")
	}
	if fv.Sparse["tags"]["travel"] != 1.0 {
		t.Error("travel tag not found")
	}
	if fv.Sparse["tags"]["music"] != 1.0 {
		t.Error("music tag not found")
	}

	// Check metadata
	if fv.Metadata["gender"] != domain.GenderFemale {
		t.Errorf("gender metadata = %v, want %v", fv.Metadata["gender"], domain.GenderFemale)
	}
	if fv.Metadata["nickname"] != "Alice" {
		t.Errorf("nickname metadata = %v, want Alice", fv.Metadata["nickname"])
	}
}

func TestNormalizeAge(t *testing.T) {
	tests := []struct {
		name string
		age  int
		want float64
	}{
		{"minimum age", 18, 0.0},
		{"maximum age", 80, 1.0},
		{"middle age", 49, (49.0 - 18.0) / (80.0 - 18.0)},
		{"below minimum", 15, 0.0},
		{"above maximum", 85, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeAge(tt.age)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("normalizeAge(%v) = %v, want %v", tt.age, got, tt.want)
			}
		})
	}
}

func TestNormalizeHeight(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   float64
	}{
		{"minimum height", 140, 0.0},
		{"maximum height", 200, 1.0},
		{"middle height", 170, (170.0 - 140.0) / (200.0 - 140.0)},
		{"below minimum", 130, 0.0},
		{"above maximum", 210, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeHeight(tt.height)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("normalizeHeight(%v) = %v, want %v", tt.height, got, tt.want)
			}
		})
	}
}

func TestNormalizeIncomeLevel(t *testing.T) {
	tests := []struct {
		name        string
		incomeLevel int
		want        float64
	}{
		{"minimum income", 0, 0.0},
		{"maximum income", 2000, 1.0},
		{"middle income", 400, 0.2},
		{"high income", 1000, 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeIncomeLevel(tt.incomeLevel)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("normalizeIncomeLevel(%v) = %v, want %v", tt.incomeLevel, got, tt.want)
			}
		})
	}
}

func TestNormalizeElo(t *testing.T) {
	tests := []struct {
		name string
		elo  int
		want float64
	}{
		{"minimum elo", 0, 0.0},
		{"maximum elo", 2000, 1.0},
		{"middle elo", 1000, 0.5},
		{"above maximum", 2500, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeElo(tt.elo)
			if math.Abs(got-tt.want) > 0.01 {
				t.Errorf("normalizeElo(%v) = %v, want %v", tt.elo, got, tt.want)
			}
		})
	}
}

func TestNormalizeActivity(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		lastActive time.Time
		want       float64
	}{
		{"active now", now, 1.0},
		{"active 15 days ago", now.AddDate(0, 0, -15), 0.5},
		{"active 30 days ago", now.AddDate(0, 0, -30), 0.0},
		{"active 60 days ago", now.AddDate(0, 0, -60), 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeActivity(tt.lastActive)
			if math.Abs(got-tt.want) > 0.05 {
				t.Errorf("normalizeActivity(%v) = %v, want %v", tt.lastActive, got, tt.want)
			}
		})
	}
}

func TestNormalizeRecency(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		createdAt time.Time
		want      float64
	}{
		{"just created", now, 1.0},
		{"6 months old", now.AddDate(0, -6, 0), 0.5},
		{"1 year old", now.AddDate(-1, 0, 0), 0.0},
		{"2 years old", now.AddDate(-2, 0, 0), 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeRecency(tt.createdAt)
			if math.Abs(got-tt.want) > 0.05 {
				t.Errorf("normalizeRecency(%v) = %v, want %v", tt.createdAt, got, tt.want)
			}
		})
	}
}

func TestNewDatingFeatureMapper(t *testing.T) {
	mapper := NewDatingFeatureMapper()
	if mapper == nil {
		t.Error("NewDatingFeatureMapper() returned nil")
	}
}
