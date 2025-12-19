package loaders

import (
	"testing"
	"time"

	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/stretchr/testify/assert"
)

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name      string
		birthdate time.Time
		expected  int
	}{
		{
			name:      "adult born 30 years ago",
			birthdate: time.Now().AddDate(-30, 0, 0),
			expected:  30,
		},
		{
			name:      "child born 10 years ago",
			birthdate: time.Now().AddDate(-10, 0, 0),
			expected:  10,
		},
		{
			name:      "exactly 13 years old",
			birthdate: time.Now().AddDate(-13, 0, 0),
			expected:  13,
		},
		{
			name:      "12 years old (under 13)",
			birthdate: time.Now().AddDate(-12, 0, 0),
			expected:  12,
		},
		{
			name:      "birthday not yet occurred this year",
			birthdate: time.Now().AddDate(-20, 0, 1), // tomorrow, 20 years ago
			expected:  19,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			age := calculateAge(tt.birthdate)
			assert.Equal(t, tt.expected, age)
		})
	}
}

func TestBuildConsentStatus_FiltersLocalConsentsForUsersUnder13(t *testing.T) {
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Create consents: one LOCAL, one REMOTE
	localConsent := &model.Consent{
		ID:             "CN01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Key:            "internal_leaderboard",
		Version:        1,
		Title:          "Internal Leaderboard",
		ManagementType: model.ConsentManagementTypeLocal,
	}
	remoteConsent := &model.Consent{
		ID:             "CN02ARZ3NDEKTSV4RRFFQ69G5FAV",
		Key:            "bcc_media_leaderboard",
		Version:        1,
		Title:          "BCC Media Leaderboard",
		ManagementType: model.ConsentManagementTypeRemote,
	}

	latestConsents := map[string]*model.Consent{
		localConsent.ID:  localConsent,
		remoteConsent.ID: remoteConsent,
	}
	userConsentsMap := make(map[string]map[string]*model.UserConsent)

	t.Run("user under 13 should not see LOCAL consents", func(t *testing.T) {
		birthdate := time.Now().AddDate(-12, 0, 0) // 12 years old
		status := buildConsentStatus(userID, birthdate, latestConsents, userConsentsMap)

		// Should only have REMOTE consent in pending
		assert.Len(t, status.PendingConsents, 1)
		assert.Equal(t, remoteConsent.Key, status.PendingConsents[0].Key)
		assert.Empty(t, status.AcceptedConsents)
		assert.Empty(t, status.RejectedConsents)
	})

	t.Run("user 13 or older should see all consents", func(t *testing.T) {
		birthdate := time.Now().AddDate(-13, 0, 0) // exactly 13 years old
		status := buildConsentStatus(userID, birthdate, latestConsents, userConsentsMap)

		// Should have both consents in pending
		assert.Len(t, status.PendingConsents, 2)
		assert.Empty(t, status.AcceptedConsents)
		assert.Empty(t, status.RejectedConsents)
	})

	t.Run("adult user should see all consents", func(t *testing.T) {
		birthdate := time.Now().AddDate(-25, 0, 0) // 25 years old
		status := buildConsentStatus(userID, birthdate, latestConsents, userConsentsMap)

		// Should have both consents in pending
		assert.Len(t, status.PendingConsents, 2)
		assert.Empty(t, status.AcceptedConsents)
		assert.Empty(t, status.RejectedConsents)
	})
}

func TestBuildConsentStatus_FiltersAcceptedAndRejectedLocalConsentsForUsersUnder13(t *testing.T) {
	userID := "US01ARZ3NDEKTSV4RRFFQ69G5FAV"

	// Create consents: one LOCAL accepted, one REMOTE rejected
	localConsent := &model.Consent{
		ID:             "CN01ARZ3NDEKTSV4RRFFQ69G5FAV",
		Key:            "internal_leaderboard",
		Version:        1,
		Title:          "Internal Leaderboard",
		ManagementType: model.ConsentManagementTypeLocal,
	}
	remoteConsent := &model.Consent{
		ID:             "CN02ARZ3NDEKTSV4RRFFQ69G5FAV",
		Key:            "bcc_media_leaderboard",
		Version:        1,
		Title:          "BCC Media Leaderboard",
		ManagementType: model.ConsentManagementTypeRemote,
	}

	latestConsents := map[string]*model.Consent{
		localConsent.ID:  localConsent,
		remoteConsent.ID: remoteConsent,
	}

	// User has accepted local consent and rejected remote consent
	userConsentsMap := map[string]map[string]*model.UserConsent{
		userID: {
			localConsent.Key: {
				ID:        "UC01ARZ3NDEKTSV4RRFFQ69G5FAV",
				ConsentID: localConsent.ID,
				Action:    model.ConsentActionAccepted,
			},
			remoteConsent.Key: {
				ID:        "UC02ARZ3NDEKTSV4RRFFQ69G5FAV",
				ConsentID: remoteConsent.ID,
				Action:    model.ConsentActionRejected,
			},
		},
	}

	t.Run("user under 13 should not see accepted LOCAL consent", func(t *testing.T) {
		birthdate := time.Now().AddDate(-12, 0, 0) // 12 years old
		status := buildConsentStatus(userID, birthdate, latestConsents, userConsentsMap)

		// Should not have local consent in accepted list (filtered out)
		assert.Empty(t, status.PendingConsents)
		assert.Empty(t, status.AcceptedConsents) // LOCAL filtered out
		assert.Len(t, status.RejectedConsents, 1)
		assert.Equal(t, remoteConsent.ID, status.RejectedConsents[0].ConsentID)
	})

	t.Run("user 13 or older should see all consents", func(t *testing.T) {
		birthdate := time.Now().AddDate(-13, 0, 0) // exactly 13 years old
		status := buildConsentStatus(userID, birthdate, latestConsents, userConsentsMap)

		// Should have both consents
		assert.Empty(t, status.PendingConsents)
		assert.Len(t, status.AcceptedConsents, 1)
		assert.Len(t, status.RejectedConsents, 1)
	})
}
