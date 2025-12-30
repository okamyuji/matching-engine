package domain

import "testing"

func TestProfile_Creation(t *testing.T) {
	tests := []struct {
		name    string
		profile Profile
	}{
		{
			name: "valid profile with all fields",
			profile: Profile{
				UserID:         "user123",
				Height:         170,
				BodyType:       BodyTypeAverage,
				Education:      EducationUniversity,
				Occupation:     "Engineer",
				IncomeLevel:    500,
				MarriageDesire: MarriageWantEventually,
				ChildrenDesire: ChildrenWant,
				Smoking:        SmokingNonSmoker,
				Drinking:       DrinkingSocial,
				Tags: []ProfileTag{
					{UserID: "user123", Tag: "sports"},
					{UserID: "user123", Tag: "travel"},
					{UserID: "user123", Tag: "music"},
				},
				SelfIntroduction: "Hello, I'm looking for a long-term relationship.",
			},
		},
		{
			name: "minimal profile",
			profile: Profile{
				UserID: "user456",
				Height: 165,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.profile

			if p.UserID == "" {
				t.Error("UserID should not be empty")
			}

			if tt.name == "valid profile with all fields" {
				if len(p.Tags) != 3 {
					t.Errorf("Tags length = %v, want 3", len(p.Tags))
				}
			}
		})
	}
}
