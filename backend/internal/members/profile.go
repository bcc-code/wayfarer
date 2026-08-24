package members

import "time"

// ProfileFields holds profile data derived from a Member record, ready to
// apply to a Wayfarer user. An empty string field or a nil Birthdate means
// the Members API had no value for it — callers should leave the existing
// database value untouched rather than overwrite it.
type ProfileFields struct {
	Email       string
	FirstName   string
	LastName    string
	MiddleName  string
	DisplayName string
	Name        string
	Birthdate   *time.Time
}

// ExtractProfile derives normalized profile fields from a member record.
func ExtractProfile(member *Member) ProfileFields {
	fields := ProfileFields{
		Email:       member.Email,
		FirstName:   member.FirstName,
		LastName:    member.LastName,
		MiddleName:  member.MiddleName,
		DisplayName: member.DisplayName,
	}

	switch {
	case fields.FirstName != "" && fields.LastName != "":
		fields.Name = fields.FirstName + " " + fields.LastName
	case fields.FirstName != "":
		fields.Name = fields.FirstName
	case fields.DisplayName != "":
		fields.Name = fields.DisplayName
	}

	if fields.DisplayName == "" {
		fields.DisplayName = GenerateDisplayName(fields.FirstName, fields.LastName, fields.Name)
	}

	if member.BirthDate != "" {
		if parsed, ok := ParseBirthdate(member.BirthDate); ok {
			fields.Birthdate = &parsed
		}
	}

	return fields
}

// GenerateDisplayName creates a display name in the format "FirstName L." if
// both names are provided, otherwise returns the fallback name.
func GenerateDisplayName(firstName, lastName, fallbackName string) string {
	if firstName != "" && lastName != "" {
		return firstName + " " + string([]rune(lastName)[0]) + "."
	}
	return fallbackName
}

// ParseBirthdate parses a Members API birthdate string (YYYY-MM-DD) and
// validates it falls within a reasonable range (1900-01-01 to today).
func ParseBirthdate(raw string) (time.Time, bool) {
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return time.Time{}, false
	}

	minDate := time.Date(1900, 1, 1, 0, 0, 0, 0, time.UTC)
	if parsed.Before(minDate) || parsed.After(time.Now()) {
		return time.Time{}, false
	}

	return parsed, true
}
