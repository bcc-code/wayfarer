package members

import (
	"github.com/google/uuid"
	"time"
)

type result[t any] struct {
	Data t `json:"data"`
}

// Member is a member with related data
type Member struct {
	PersonID      int
	Uid           uuid.UUID `json:"uid"`
	BirthDate     string
	Email         string
	EmailVerified bool   `json:"emailVerified"`
	DisplayName   string `json:"displayName"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	MiddleName    string `json:"middleName"`
	Gender        string `json:"gender"`
	Affiliations  []Affiliation
}

// Affiliation is an affiliation to an entity
type Affiliation struct {
	Active    bool       `json:"isActive"`
	OrgUid    uuid.UUID  `json:"orgUid"`
	PersonUid uuid.UUID  `json:"personUid"`
	Uid       uuid.UUID  `json:"uid"`
	Type      string     `json:"type"`
	ValidFrom *time.Time `json:"validFrom"`
	ValidTo   *time.Time `json:"validTo"`
}

// IsActive returns true if the affiliation is currently active.
// Checks: Active flag, ValidFrom <= now, ValidTo > now (or nil)
func (a Affiliation) IsActive() bool {
	if !a.Active {
		return false
	}
	now := time.Now()
	if a.ValidFrom != nil && now.Before(*a.ValidFrom) {
		return false
	}
	if a.ValidTo != nil && now.After(*a.ValidTo) {
		return false
	}
	return true
}

// FilterActiveAffiliations returns only the affiliations that are currently active.
func FilterActiveAffiliations(affiliations []Affiliation) []Affiliation {
	var result []Affiliation
	for _, aff := range affiliations {
		if aff.IsActive() {
			result = append(result, aff)
		}
	}
	return result
}

// GetActiveAffiliationOrgUIDs returns the OrgUids of all currently active affiliations.
func GetActiveAffiliationOrgUIDs(affiliations []Affiliation) []uuid.UUID {
	var result []uuid.UUID
	for _, aff := range affiliations {
		if aff.IsActive() {
			result = append(result, aff.OrgUid)
		}
	}
	return result
}

// HasActiveAffiliation returns true if there is at least one active affiliation.
func HasActiveAffiliation(affiliations []Affiliation) bool {
	for _, aff := range affiliations {
		if aff.IsActive() {
			return true
		}
	}
	return false
}

// RoleAssignment represents a role assignment for a member
type RoleAssignment struct {
	Uid     uuid.UUID `json:"uid"`
	RoleUid uuid.UUID `json:"roleUid"`
	OrgUid  uuid.UUID `json:"orgUid"`
}

// MemberWithRoles is a member with role assignments
type MemberWithRoles struct {
	PersonID        int              `json:"personID"`
	Uid             uuid.UUID        `json:"uid"`
	BirthDate       string           `json:"birthDate"`
	Email           string           `json:"email"`
	EmailVerified   bool             `json:"emailVerified"`
	DisplayName     string           `json:"displayName"`
	FirstName       string           `json:"firstName"`
	LastName        string           `json:"lastName"`
	MiddleName      string           `json:"middleName"`
	Gender          string           `json:"gender"`
	RoleAssignments []RoleAssignment `json:"roleAssignments"`
}

// OrganizationAddress contains address data for an organization
type OrganizationAddress struct {
	CountryCode string `json:"countryCode"`
}

// Organization contains organizational data
type Organization struct {
	OrgID           int
	Name            string `json:"districtName"`
	Type            string
	Uid             uuid.UUID
	VisitingAddress OrganizationAddress `json:"visitingAddress"`
}
