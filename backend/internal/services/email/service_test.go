package email

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildFeedbackEmailBody(t *testing.T) {
	t.Run("full email with all fields", func(t *testing.T) {
		service := &Service{
			adminBaseURL: "https://admin.example.com",
		}

		params := FeedbackEmailParams{
			UserID:      "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			UserName:    "John Doe",
			UserEmail:   "john@example.com",
			ChurchName:  "Oslo Church",
			TotalPoints: 150,
			ProjectName: "Summer Study 2024",
			UserCreated: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
			Consents: []ConsentEntry{
				{Title: "Privacy Policy", Action: "Accepted", Date: time.Date(2023, 6, 15, 10, 0, 0, 0, time.UTC)},
				{Title: "Marketing", Action: "Rejected", Date: time.Date(2023, 6, 16, 10, 0, 0, 0, time.UTC)},
			},
			Message: "This is my feedback message.",
		}

		body := service.buildFeedbackEmailBody(params)

		assert.Contains(t, body, "User: US01ARZ3NDEKTSV4RRFFQ69G5FAV (https://admin.example.com/admin/users/US01ARZ3NDEKTSV4RRFFQ69G5FAV)")
		assert.Contains(t, body, "Church: Oslo Church")
		assert.Contains(t, body, "Total Points: 150 (in Summer Study 2024)")
		assert.Contains(t, body, "Created: 2023-06-15")
		assert.Contains(t, body, "Privacy Policy: Accepted on 2023-06-15")
		assert.Contains(t, body, "Marketing: Rejected on 2023-06-16")
		assert.Contains(t, body, "Message:")
		assert.Contains(t, body, "This is my feedback message.")
	})

	t.Run("without admin base URL", func(t *testing.T) {
		service := &Service{
			adminBaseURL: "",
		}

		params := FeedbackEmailParams{
			UserID:      "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			UserName:    "John Doe",
			UserEmail:   "john@example.com",
			ChurchName:  "Oslo Church",
			TotalPoints: 0,
			ProjectName: "",
			UserCreated: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
			Consents:    []ConsentEntry{},
			Message:     "Feedback without project context.",
		}

		body := service.buildFeedbackEmailBody(params)

		// Should just show the user ID without a link
		assert.Contains(t, body, "User: US01ARZ3NDEKTSV4RRFFQ69G5FAV")
		assert.NotContains(t, body, "https://")
	})

	t.Run("without project context", func(t *testing.T) {
		service := &Service{
			adminBaseURL: "https://admin.example.com",
		}

		params := FeedbackEmailParams{
			UserID:      "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			UserName:    "Jane Doe",
			UserEmail:   "jane@example.com",
			ChurchName:  "Bergen Church",
			TotalPoints: 0,
			ProjectName: "",
			UserCreated: time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
			Consents:    []ConsentEntry{},
			Message:     "General feedback.",
		}

		body := service.buildFeedbackEmailBody(params)

		assert.Contains(t, body, "Total Points: N/A (no project context)")
	})

	t.Run("without consents", func(t *testing.T) {
		service := &Service{
			adminBaseURL: "https://admin.example.com",
		}

		params := FeedbackEmailParams{
			UserID:      "US01ARZ3NDEKTSV4RRFFQ69G5FAV",
			UserName:    "Test User",
			UserEmail:   "test@example.com",
			ChurchName:  "Test Church",
			TotalPoints: 50,
			ProjectName: "Test Project",
			UserCreated: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Consents:    []ConsentEntry{},
			Message:     "Test message.",
		}

		body := service.buildFeedbackEmailBody(params)

		assert.Contains(t, body, "Consents:")
		assert.Contains(t, body, "No consent records")
	})
}

func TestNewService(t *testing.T) {
	t.Run("with API key", func(t *testing.T) {
		service := NewService("test-api-key", "https://admin.example.com")

		assert.NotNil(t, service)
		assert.NotNil(t, service.client)
		assert.Equal(t, "https://admin.example.com", service.adminBaseURL)
	})

	t.Run("without API key", func(t *testing.T) {
		service := NewService("", "https://admin.example.com")

		assert.NotNil(t, service)
		assert.Nil(t, service.client)
		assert.Equal(t, "https://admin.example.com", service.adminBaseURL)
	})
}

func TestConsentEntry(t *testing.T) {
	t.Run("consent entry fields", func(t *testing.T) {
		entry := ConsentEntry{
			Title:  "Terms of Service",
			Action: "Accepted",
			Date:   time.Date(2024, 3, 15, 14, 30, 0, 0, time.UTC),
		}

		assert.Equal(t, "Terms of Service", entry.Title)
		assert.Equal(t, "Accepted", entry.Action)
		assert.Equal(t, 2024, entry.Date.Year())
		assert.Equal(t, time.March, entry.Date.Month())
		assert.Equal(t, 15, entry.Date.Day())
	})
}

func TestEmailBodyStructure(t *testing.T) {
	service := &Service{adminBaseURL: "https://admin.test.com"}

	params := FeedbackEmailParams{
		UserID:      "US123",
		UserName:    "Test",
		UserEmail:   "test@test.com",
		ChurchName:  "Church",
		TotalPoints: 100,
		ProjectName: "Project",
		UserCreated: time.Now(),
		Consents: []ConsentEntry{
			{Title: "Consent1", Action: "Accepted", Date: time.Now()},
		},
		Message: "Test message",
	}

	body := service.buildFeedbackEmailBody(params)

	// Verify sections appear in correct order (Message first, then user context)
	messageIndex := strings.Index(body, "Message:")
	userIndex := strings.Index(body, "User:")
	churchIndex := strings.Index(body, "Church:")
	pointsIndex := strings.Index(body, "Total Points:")
	createdIndex := strings.Index(body, "Created:")
	consentsIndex := strings.Index(body, "Consents:")

	assert.True(t, messageIndex < userIndex, "Message should come before User")
	assert.True(t, userIndex < churchIndex, "User should come before Church")
	assert.True(t, churchIndex < pointsIndex, "Church should come before Total Points")
	assert.True(t, pointsIndex < createdIndex, "Total Points should come before Created")
	assert.True(t, createdIndex < consentsIndex, "Created should come before Consents")
}
