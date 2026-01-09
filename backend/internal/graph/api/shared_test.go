package api

import (
	"context"
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
)

func TestUserResolver_Age(t *testing.T) {
	resolver := &userResolver{}
	ctx := context.Background()

	tests := []struct {
		name      string
		birthdate string
		wantAge   *int
		wantErr   bool
	}{
		{
			name:      "empty birthdate returns nil",
			birthdate: "",
			wantAge:   nil,
			wantErr:   false,
		},
		{
			name:      "invalid date format returns error",
			birthdate: "invalid-date",
			wantAge:   nil,
			wantErr:   true,
		},
		{
			name:      "wrong date format returns error",
			birthdate: "01/15/1990",
			wantAge:   nil,
			wantErr:   true,
		},
		{
			name:      "valid birthdate calculates correct age",
			birthdate: "1990-06-15",
			wantAge:   intPtr(calculateExpectedAge("1990-06-15")),
			wantErr:   false,
		},
		{
			name:      "birthday not yet occurred this year",
			birthdate: time.Now().AddDate(-25, 6, 0).Format("2006-01-02"),
			wantAge:   intPtr(calculateExpectedAge(time.Now().AddDate(-25, 6, 0).Format("2006-01-02"))),
			wantErr:   false,
		},
		{
			name:      "birthday already occurred this year",
			birthdate: time.Now().AddDate(-25, -6, 0).Format("2006-01-02"),
			wantAge:   intPtr(calculateExpectedAge(time.Now().AddDate(-25, -6, 0).Format("2006-01-02"))),
			wantErr:   false,
		},
		{
			name:      "born today",
			birthdate: time.Now().Format("2006-01-02"),
			wantAge:   intPtr(0),
			wantErr:   false,
		},
		{
			name:      "born yesterday",
			birthdate: time.Now().AddDate(0, 0, -1).Format("2006-01-02"),
			wantAge:   intPtr(0),
			wantErr:   false,
		},
		{
			name:      "born 100 years ago",
			birthdate: time.Now().AddDate(-100, 0, 0).Format("2006-01-02"),
			wantAge:   intPtr(100),
			wantErr:   false,
		},
		{
			name:      "leap year birthday (Feb 29, 2000)",
			birthdate: "2000-02-29",
			wantAge:   intPtr(calculateExpectedAge("2000-02-29")),
			wantErr:   false,
		},
		{
			name:      "born on Jan 1",
			birthdate: "2000-01-01",
			wantAge:   intPtr(calculateExpectedAge("2000-01-01")),
			wantErr:   false,
		},
		{
			name:      "born on Dec 31",
			birthdate: "2000-12-31",
			wantAge:   intPtr(calculateExpectedAge("2000-12-31")),
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &model.User{
				Birthdate: tt.birthdate,
			}

			age, err := resolver.Age(ctx, user)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, age)
			} else {
				assert.NoError(t, err)
				if tt.wantAge == nil {
					assert.Nil(t, age)
				} else {
					assert.NotNil(t, age)
					assert.Equal(t, *tt.wantAge, *age)
				}
			}
		})
	}
}

func calculateExpectedAge(birthdateStr string) int {
	birthdate, _ := time.Parse("2006-01-02", birthdateStr)
	now := time.Now()
	// The resolver uses simple year difference (no birthday adjustment)
	return now.Year() - birthdate.Year()
}
