package domain

import "testing"

func TestPreference_Creation(t *testing.T) {
	tests := []struct {
		name       string
		preference Preference
	}{
		{
			name: "valid preference with all criteria",
			preference: Preference{
				UserID:    "user123",
				AgeMin:    25,
				AgeMax:    35,
				HeightMin: 160,
				HeightMax: 180,
				IncomeMin: 400,
				Prefectures: []PreferencePrefecture{
					{UserID: "user123", Prefecture: PrefectureTokyo},
					{UserID: "user123", Prefecture: PrefectureOsaka},
				},
			},
		},
		{
			name: "minimal preference with only age",
			preference: Preference{
				UserID: "user456",
				AgeMin: 20,
				AgeMax: 40,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.preference

			if p.UserID == "" {
				t.Error("UserID should not be empty")
			}

			if tt.name == "valid preference with all criteria" {
				if len(p.Prefectures) != 2 {
					t.Errorf("Prefectures length = %v, want 2", len(p.Prefectures))
				}
				if p.AgeMin > p.AgeMax {
					t.Error("AgeMin should not be greater than AgeMax")
				}
				if p.HeightMin > p.HeightMax {
					t.Error("HeightMin should not be greater than HeightMax")
				}
			}
		})
	}
}

func TestPreference_GetPrefectureStrings(t *testing.T) {
	p := Preference{
		UserID: "user123",
		Prefectures: []PreferencePrefecture{
			{UserID: "user123", Prefecture: PrefectureTokyo},
			{UserID: "user123", Prefecture: PrefectureOsaka},
			{UserID: "user123", Prefecture: PrefectureKyoto},
		},
	}

	result := p.GetPrefectureStrings()

	if len(result) != 3 {
		t.Errorf("GetPrefectureStrings() returned %d prefectures, want 3", len(result))
	}

	expectedPrefectures := map[string]bool{
		"tokyo": true,
		"osaka": true,
		"kyoto": true,
	}

	for _, pref := range result {
		if !expectedPrefectures[pref] {
			t.Errorf("Unexpected prefecture: %s", pref)
		}
	}
}

func TestPreference_GetEducationStrings(t *testing.T) {
	p := Preference{
		UserID: "user123",
		Educations: []PreferenceEducation{
			{UserID: "user123", Education: EducationUniversity},
			{UserID: "user123", Education: EducationGraduate},
		},
	}

	result := p.GetEducationStrings()

	if len(result) != 2 {
		t.Errorf("GetEducationStrings() returned %d educations, want 2", len(result))
	}

	expectedEducations := map[string]bool{
		"university": true,
		"graduate":   true,
	}

	for _, edu := range result {
		if !expectedEducations[edu] {
			t.Errorf("Unexpected education: %s", edu)
		}
	}
}

func TestPreference_GetMarriageDesireStrings(t *testing.T) {
	p := Preference{
		UserID: "user123",
		MarriageDesires: []PreferenceMarriageDesire{
			{UserID: "user123", MarriageDesire: MarriageWantSoon},
			{UserID: "user123", MarriageDesire: MarriageWantEventually},
		},
	}

	result := p.GetMarriageDesireStrings()

	if len(result) != 2 {
		t.Errorf("GetMarriageDesireStrings() returned %d desires, want 2", len(result))
	}

	expectedDesires := map[string]bool{
		"want_soon":       true,
		"want_eventually": true,
	}

	for _, desire := range result {
		if !expectedDesires[desire] {
			t.Errorf("Unexpected marriage desire: %s", desire)
		}
	}
}

func TestPreference_GetSmokingStatusStrings(t *testing.T) {
	p := Preference{
		UserID: "user123",
		SmokingStatuses: []PreferenceSmokingStatus{
			{UserID: "user123", SmokingStatus: SmokingNonSmoker},
			{UserID: "user123", SmokingStatus: SmokingOccasional},
		},
	}

	result := p.GetSmokingStatusStrings()

	if len(result) != 2 {
		t.Errorf("GetSmokingStatusStrings() returned %d statuses, want 2", len(result))
	}

	expectedStatuses := map[string]bool{
		"non_smoker": true,
		"occasional": true,
	}

	for _, status := range result {
		if !expectedStatuses[status] {
			t.Errorf("Unexpected smoking status: %s", status)
		}
	}
}

func TestPreference_GetDrinkingStatusStrings(t *testing.T) {
	p := Preference{
		UserID: "user123",
		DrinkingStatuses: []PreferenceDrinkingStatus{
			{UserID: "user123", DrinkingStatus: DrinkingSocial},
			{UserID: "user123", DrinkingStatus: DrinkingRegular},
		},
	}

	result := p.GetDrinkingStatusStrings()

	if len(result) != 2 {
		t.Errorf("GetDrinkingStatusStrings() returned %d statuses, want 2", len(result))
	}

	expectedStatuses := map[string]bool{
		"social":  true,
		"regular": true,
	}

	for _, status := range result {
		if !expectedStatuses[status] {
			t.Errorf("Unexpected drinking status: %s", status)
		}
	}
}

func TestPreference_GetEmptyCollections(t *testing.T) {
	p := Preference{
		UserID: "user123",
	}

	if len(p.GetPrefectureStrings()) != 0 {
		t.Error("GetPrefectureStrings() should return empty slice")
	}
	if len(p.GetEducationStrings()) != 0 {
		t.Error("GetEducationStrings() should return empty slice")
	}
	if len(p.GetMarriageDesireStrings()) != 0 {
		t.Error("GetMarriageDesireStrings() should return empty slice")
	}
	if len(p.GetSmokingStatusStrings()) != 0 {
		t.Error("GetSmokingStatusStrings() should return empty slice")
	}
	if len(p.GetDrinkingStatusStrings()) != 0 {
		t.Error("GetDrinkingStatusStrings() should return empty slice")
	}
}
