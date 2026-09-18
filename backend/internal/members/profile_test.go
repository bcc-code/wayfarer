package members

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBirthdate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expectOK bool
	}{
		{name: "valid date", input: "1983-01-01", expectOK: true},
		{name: "invalid format", input: "01/01/1983", expectOK: false},
		{name: "empty string", input: "", expectOK: false},
		{name: "too far in the past", input: "1899-12-31", expectOK: false},
		{name: "in the future", input: "2999-01-01", expectOK: false},
		{name: "exactly the minimum date", input: "1900-01-01", expectOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, ok := ParseBirthdate(tt.input)
			assert.Equal(t, tt.expectOK, ok)
			if tt.expectOK {
				assert.Equal(t, tt.input, parsed.Format("2006-01-02"))
			}
		})
	}
}

func TestGenerateDisplayName(t *testing.T) {
	tests := []struct {
		name         string
		firstName    string
		lastName     string
		fallbackName string
		expected     string
	}{
		{name: "both names present", firstName: "Alice", lastName: "Anderson", fallbackName: "fallback", expected: "Alice A."},
		{name: "only first name", firstName: "Alice", lastName: "", fallbackName: "fallback", expected: "fallback"},
		{name: "no names", firstName: "", lastName: "", fallbackName: "fallback", expected: "fallback"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateDisplayName(tt.firstName, tt.lastName, tt.fallbackName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractProfile(t *testing.T) {
	t.Run("full member data", func(t *testing.T) {
		member := &Member{
			Email:      "user@example.com",
			FirstName:  "Alice",
			LastName:   "Anderson",
			MiddleName: "M",
			BirthDate:  "1983-01-01",
		}

		fields := ExtractProfile(member)

		assert.Equal(t, "user@example.com", fields.Email)
		assert.Equal(t, "Alice Anderson", fields.Name)
		assert.Equal(t, "Alice A.", fields.DisplayName)
		require.NotNil(t, fields.Birthdate)
		assert.Equal(t, "1983-01-01", fields.Birthdate.Format("2006-01-02"))
	})

	t.Run("no birthdate, no first/last name, has display name", func(t *testing.T) {
		member := &Member{
			Email:       "user@example.com",
			DisplayName: "Test User",
		}

		fields := ExtractProfile(member)

		assert.Equal(t, "Test User", fields.Name)
		assert.Equal(t, "Test User", fields.DisplayName)
		assert.Nil(t, fields.Birthdate)
	})

	t.Run("invalid birthdate is dropped, not errored", func(t *testing.T) {
		member := &Member{
			Email:     "user@example.com",
			BirthDate: "not-a-date",
		}

		fields := ExtractProfile(member)

		assert.Nil(t, fields.Birthdate)
	})

	t.Run("empty member has empty fields", func(t *testing.T) {
		member := &Member{}

		fields := ExtractProfile(member)

		assert.Equal(t, "", fields.Email)
		assert.Equal(t, "", fields.Name)
		assert.Equal(t, "", fields.DisplayName)
		assert.Nil(t, fields.Birthdate)
	})
}
