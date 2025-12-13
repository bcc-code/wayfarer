package api

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/services"
)

// canAccessUser checks if currentUserID can access targetUserID's User object.
//
// Returns true if any of these conditions are met:
//   - currentUserID is the target user (self-access)
//   - currentUserID has admin or superadmin role
//   - currentUserID has M2M role (for system integration)
//
// Returns false otherwise.
//
// This function is used to protect User objects from unauthorized access
// in GraphQL field resolvers that return User or [User] types.
func canAccessUser(
	ctx context.Context,
	roleService *services.RoleService,
	currentUserID string,
	targetUserID string,
) bool {
	// Self-access always allowed
	if currentUserID == targetUserID {
		return true
	}

	// Admins and superadmins can access any user
	if roleService.IsAdmin(ctx, currentUserID) {
		return true
	}

	// M2M service accounts can access any user
	if roleService.HasRole(ctx, currentUserID, services.RoleM2M) {
		return true
	}

	return false
}
