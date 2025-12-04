package ssf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsChapterPublishedAfterDate(t *testing.T) {
	// In production, we compare against start of tomorrow to filter for "after today"
	// Here we test the raw function which checks if published > referenceDate
	referenceDate := time.Date(2025, 12, 5, 0, 0, 0, 0, time.UTC) // start of Dec 5 (tomorrow)

	tests := []struct {
		name              string
		datetimePublished string
		expected          bool
	}{
		{
			name:              "published on reference date",
			datetimePublished: "2025-12-05T00:00:00Z",
			expected:          false, // not strictly after
		},
		{
			name:              "published after reference date",
			datetimePublished: "2025-12-05T00:00:01Z",
			expected:          true,
		},
		{
			name:              "published next week",
			datetimePublished: "2025-12-11T12:00:00Z",
			expected:          true,
		},
		{
			name:              "published before reference date",
			datetimePublished: "2025-12-04T23:59:59Z",
			expected:          false,
		},
		{
			name:              "published yesterday",
			datetimePublished: "2025-12-03T12:00:00Z",
			expected:          false,
		},
		{
			name:              "published last month",
			datetimePublished: "2025-11-04T00:00:00Z",
			expected:          false,
		},
		{
			name:              "empty date",
			datetimePublished: "",
			expected:          false,
		},
		{
			name:              "invalid date format",
			datetimePublished: "not-a-date",
			expected:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chapter := &PlanChapter{
				DatetimePublished: tt.datetimePublished,
			}
			result := isChapterPublishedAfterDate(chapter, referenceDate)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParsePublishedDate(t *testing.T) {
	tests := []struct {
		name              string
		datetimePublished string
		expectNil         bool
		expectedTime      time.Time
	}{
		{
			name:              "valid RFC3339",
			datetimePublished: "2025-12-05T10:30:00Z",
			expectNil:         false,
			expectedTime:      time.Date(2025, 12, 5, 10, 30, 0, 0, time.UTC),
		},
		{
			name:              "valid with timezone offset",
			datetimePublished: "2025-12-05T10:30:00+01:00",
			expectNil:         false,
			expectedTime:      time.Date(2025, 12, 5, 10, 30, 0, 0, time.FixedZone("", 3600)),
		},
		{
			name:              "empty string",
			datetimePublished: "",
			expectNil:         true,
		},
		{
			name:              "invalid format",
			datetimePublished: "2025-12-05",
			expectNil:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePublishedDate(tt.datetimePublished)
			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.True(t, tt.expectedTime.Equal(*result), "expected %v, got %v", tt.expectedTime, *result)
			}
		})
	}
}
