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
