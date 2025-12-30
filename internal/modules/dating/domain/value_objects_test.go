package domain

import "testing"

func TestIncomeRange_Level(t *testing.T) {
	tests := []struct {
		name   string
		income IncomeRange
		want   int
	}{
		{"income 0-2M", Income0to200, 0},
		{"income 2-4M", Income200to400, 1},
		{"income 4-6M", Income400to600, 2},
		{"income 6-8M", Income600to800, 3},
		{"income 8-10M", Income800to1000, 4},
		{"income 10-15M", Income1000to1500, 5},
		{"income 15-20M", Income1500to2000, 6},
		{"income 20-30M", Income2000to3000, 7},
		{"income 30-50M", Income3000to5000, 8},
		{"income 50M+", Income5000Plus, 9},
		{"income private", IncomePrivate, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.income.Level(); got != tt.want {
				t.Errorf("Level() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGender_Values(t *testing.T) {
	genders := []Gender{GenderMale, GenderFemale, GenderOther}

	if len(genders) != 3 {
		t.Errorf("Expected 3 gender values, got %d", len(genders))
	}

	if GenderMale != "male" {
		t.Errorf("GenderMale = %v, want 'male'", GenderMale)
	}
	if GenderFemale != "female" {
		t.Errorf("GenderFemale = %v, want 'female'", GenderFemale)
	}
	if GenderOther != "other" {
		t.Errorf("GenderOther = %v, want 'other'", GenderOther)
	}
}

func TestPrefecture_Values(t *testing.T) {
	prefectures := []Prefecture{
		PrefectureTokyo,
		PrefectureOsaka,
		PrefectureKyoto,
		PrefectureAichi,
		PrefectureFukuoka,
	}

	if len(prefectures) != 5 {
		t.Errorf("Expected 5 prefecture values, got %d", len(prefectures))
	}

	if PrefectureTokyo != "tokyo" {
		t.Errorf("PrefectureTokyo = %v, want 'tokyo'", PrefectureTokyo)
	}
}

func TestBodyType_Values(t *testing.T) {
	bodyTypes := []BodyType{
		BodyTypeSlim,
		BodyTypeAverage,
		BodyTypeAthletic,
		BodyTypeLarge,
	}

	if len(bodyTypes) != 4 {
		t.Errorf("Expected 4 body type values, got %d", len(bodyTypes))
	}
}

func TestEducation_Values(t *testing.T) {
	educations := []Education{
		EducationHighSchool,
		EducationVocational,
		EducationUniversity,
		EducationGraduate,
	}

	if len(educations) != 4 {
		t.Errorf("Expected 4 education values, got %d", len(educations))
	}
}

func TestMarriageDesire_Values(t *testing.T) {
	desires := []MarriageDesire{
		MarriageWantSoon,
		MarriageWantEventually,
		MarriageUndecided,
		MarriageNotWant,
	}

	if len(desires) != 4 {
		t.Errorf("Expected 4 marriage desire values, got %d", len(desires))
	}
}

func TestChildrenDesire_Values(t *testing.T) {
	desires := []ChildrenDesire{
		ChildrenWant,
		ChildrenNotWant,
		ChildrenUndecided,
	}

	if len(desires) != 3 {
		t.Errorf("Expected 3 children desire values, got %d", len(desires))
	}
}

func TestSmokingStatus_Values(t *testing.T) {
	statuses := []SmokingStatus{
		SmokingNonSmoker,
		SmokingOccasional,
		SmokingSmoker,
	}

	if len(statuses) != 3 {
		t.Errorf("Expected 3 smoking status values, got %d", len(statuses))
	}
}

func TestDrinkingStatus_Values(t *testing.T) {
	statuses := []DrinkingStatus{
		DrinkingNonDrinker,
		DrinkingSocial,
		DrinkingRegular,
	}

	if len(statuses) != 3 {
		t.Errorf("Expected 3 drinking status values, got %d", len(statuses))
	}
}

func TestPhoto_Creation(t *testing.T) {
	photo := Photo{
		URL:       "https://example.com/photo.jpg",
		IsPrimary: true,
		Order:     1,
	}

	if photo.URL == "" {
		t.Error("Photo URL should not be empty")
	}
	if !photo.IsPrimary {
		t.Error("Photo should be primary")
	}
	if photo.Order != 1 {
		t.Errorf("Photo order = %v, want 1", photo.Order)
	}
}
