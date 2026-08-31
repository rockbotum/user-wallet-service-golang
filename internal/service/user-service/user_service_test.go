package user_service

import (
	"testing"
	"time"
)

func TestCalculateAge(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		birthDay time.Time
		wantAge  int
	}{
		{
			name:     "birthday already passed this year",
			birthDay: now.AddDate(-30, 0, -1),
			wantAge:  30,
		},
		{
			name:     "birthday not yet this year",
			birthDay: now.AddDate(-30, 0, 1),
			wantAge:  29,
		},
		{
			name:     "birthday today",
			birthDay: now.AddDate(-25, 0, 0),
			wantAge:  25,
		},
		{
			name:     "exact year boundary",
			birthDay: now.AddDate(-1, 0, 0),
			wantAge:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateAge(tt.birthDay); got != tt.wantAge {
				t.Fatalf("calculateAge() = %d, want %d", got, tt.wantAge)
			}
		})
	}
}
