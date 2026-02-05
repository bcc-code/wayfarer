package members

import "strings"

// NormalizeGender converts gender string to database format (MALE, FEMALE, UNKNOWN)
func NormalizeGender(gender string) string {
	gender = strings.ToUpper(strings.TrimSpace(gender))
	if gender == "MALE" || gender == "M" {
		return "MALE"
	}
	if gender == "FEMALE" || gender == "F" {
		return "FEMALE"
	}
	return "UNKNOWN"
}
