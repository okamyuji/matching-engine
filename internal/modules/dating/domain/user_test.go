package domain

import (
	"testing"
	"time"
)

func TestUser_Age(t *testing.T) {
	tests := []struct {
		name      string
		birthDate time.Time
		wantAge   int
	}{
		{
			name:      "age 25 - birthday already passed this year",
			birthDate: time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC),
			wantAge:   25,
		},
		{
			name:      "age 30 - birthday not yet this year",
			birthDate: time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC),
			wantAge:   30,
		},
		{
			name:      "age 18 - just turned 18",
			birthDate: time.Now().AddDate(-18, 0, 0),
			wantAge:   18,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &User{
				BirthDate: tt.birthDate,
			}

			got := u.Age()

			// For the dynamic test case, we need to check if the age is correct
			// based on the current date
			if tt.name == "age 18 - just turned 18" {
				if got != 18 {
					t.Errorf("Age() = %v, want %v", got, 18)
				}
				return
			}

			// For fixed dates, calculate expected age dynamically
			now := time.Now()
			expectedAge := now.Year() - tt.birthDate.Year()
			if now.Month() < tt.birthDate.Month() ||
				(now.Month() == tt.birthDate.Month() && now.Day() < tt.birthDate.Day()) {
				expectedAge--
			}

			if got != expectedAge {
				t.Errorf("Age() = %v, want %v", got, expectedAge)
			}
		})
	}
}

func TestUser_AgeBeforeBirthday(t *testing.T) {
	// Test a user whose birthday hasn't occurred this year
	birthDate := time.Date(2000, time.December, 31, 0, 0, 0, 0, time.UTC)
	u := &User{
		BirthDate: birthDate,
	}

	age := u.Age()

	// Calculate expected age
	now := time.Now()
	expectedAge := now.Year() - 2000
	if now.Month() < time.December ||
		(now.Month() == time.December && now.Day() < 31) {
		expectedAge--
	}

	if age != expectedAge {
		t.Errorf("Age() = %v, want %v", age, expectedAge)
	}
}
