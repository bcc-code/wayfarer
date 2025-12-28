package members

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/utils"
	"github.com/google/uuid"
)

const (
	personFields = "*,affiliations.*"
	orgFields    = "districtName,type,orgID,uid,visitingAddress.countryCode"
)

// Lookup returns a member from the members api
func (c *Client) Lookup(ctx context.Context, personID int) (*Member, error) {
	return get[Member](ctx, c, fmt.Sprintf("v2/persons/%d?fields=%s", personID, personFields))
}

// RetrieveByEmails retrieves members by emails
func (c *Client) RetrieveByEmails(ctx context.Context, emails []string) (*[]Member, error) {
	filter := map[string]any{
		"email": map[string]any{
			"_in": emails,
		},
	}

	encoded, _ := json.Marshal(filter)

	return get[[]Member](ctx, c, fmt.Sprintf("v2/persons?filter=%s&fields=%s", encoded, personFields))
}

// GetMembersByIDs retrieves a batch of members by ID
func (c *Client) GetMembersByIDs(ctx context.Context, ids []int) ([]Member, error) {
	chunkedIds := utils.Chunk(ids, 800)
	var out []Member

	for _, chunk := range chunkedIds {
		filter := map[string]any{
			"personID": map[string]any{
				"_in": chunk,
			},
		}

		encoded, _ := json.Marshal(filter)

		ms, err := get[[]Member](ctx, c, fmt.Sprintf("v2/persons?limit=999&filter=%s&fields=%s", encoded, personFields))
		if err != nil {
			return nil, fmt.Errorf("failed to get members chunk: %w", err)
		}

		out = append(out, *ms...)
	}

	return out, nil
}

// GetOrganizationByOrgID returns an organization by its OrgID (external_id).
func (c *Client) GetOrganizationByOrgID(ctx context.Context, orgID int) (*Organization, error) {
	filter := map[string]any{
		"orgID": map[string]any{
			"_eq": orgID,
		},
	}

	encoded, _ := json.Marshal(filter)

	orgs, err := get[[]Organization](ctx, c, fmt.Sprintf("v2/orgs?limit=1&filter=%s&fields=%s", encoded, orgFields))
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if len(*orgs) == 0 {
		return nil, fmt.Errorf("organization with orgID %d not found", orgID)
	}

	return &(*orgs)[0], nil
}

// GetOrganizationByUID returns an organization by its UUID.
func (c *Client) GetOrganizationByUID(ctx context.Context, uid uuid.UUID) (*Organization, error) {
	filter := map[string]any{
		"uid": map[string]any{
			"_eq": uid.String(),
		},
	}

	encoded, _ := json.Marshal(filter)

	orgs, err := get[[]Organization](ctx, c, fmt.Sprintf("v2/orgs?limit=1&filter=%s&fields=%s", encoded, orgFields))
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if len(*orgs) == 0 {
		return nil, fmt.Errorf("organization with uid %s not found", uid.String())
	}

	return &(*orgs)[0], nil
}

// GetOrganizationsByIDs returns organizations by IDs.
func (c *Client) GetOrganizationsByIDs(ctx context.Context, ids []uuid.UUID) ([]Organization, error) {
	chunkedIds := utils.Chunk(ids, 800)
	var out []Organization

	for _, chunk := range chunkedIds {
		filter := map[string]any{
			"uid": map[string]any{
				"_in": chunk,
			},
		}

		encoded, _ := json.Marshal(filter)

		ms, err := get[[]Organization](ctx, c, fmt.Sprintf("v2/orgs?limit=999&filter=%s&fields=%s", encoded, orgFields))
		if err != nil {
			return nil, fmt.Errorf("failed to get organizations chunk: %w", err)
		}

		out = append(out, *ms...)
	}

	return out, nil
}

// GetAllOrganizations returns all organizations from the Members API.
func (c *Client) GetAllOrganizations(ctx context.Context) ([]Organization, error) {
	orgs, err := get[[]Organization](ctx, c, fmt.Sprintf("v2/orgs?limit=999&fields=%s", orgFields))
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	return *orgs, nil
}
