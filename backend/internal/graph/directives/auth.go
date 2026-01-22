package directives

import (
	"context"
	"fmt"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/graph-gophers/dataloader/v7"
)

// NewRequireRole creates a RequireRole directive that looks up roles from the database
// via a cached dataloader for regular users, while M2M users continue using token-based roles.
func NewRequireRole(rolesLoader *dataloader.Loader[string, []*model.UserRole]) func(ctx context.Context, obj any, next graphql.Resolver, roles []string) (any, error) {
	return func(ctx context.Context, obj any, next graphql.Resolver, roles []string) (any, error) {
		// Get token roles from context
		tokenRoles, ok := ctx.Value(middleware.UserRolesKey).([]string)
		if !ok || len(tokenRoles) == 0 {
			return nil, fmt.Errorf("unauthorized: no roles found in context")
		}

		// Check if M2M user (bypass DB lookup)
		isM2M := false
		for _, role := range tokenRoles {
			if role == "m2m" {
				isM2M = true
				break
			}
		}

		// M2M: use token roles
		if isM2M {
			for _, userRole := range tokenRoles {
				for _, allowedRole := range roles {
					if userRole == allowedRole {
						return next(ctx)
					}
				}
			}
			return nil, fmt.Errorf("unauthorized: M2M roles %v do not match required roles %v", tokenRoles, roles)
		}

		// Non-M2M: lookup roles from database via dataloader (cached)
		userID, ok := middleware.GetUserID(ctx)
		if !ok || userID == "" {
			return nil, fmt.Errorf("unauthorized: user ID not found in context")
		}

		// Use the cached dataloader to get roles
		dbRoles, err := rolesLoader.Load(ctx, userID)()
		if err != nil {
			return nil, fmt.Errorf("unauthorized: failed to load user roles: %w", err)
		}

		// Build effective roles from DB
		// All authenticated users implicitly have the "user" role
		effectiveRoles := []string{"user"}
		for _, dbRole := range dbRoles {
			effectiveRoles = append(effectiveRoles, strings.ToLower(string(dbRole.Role)))
		}

		// Check if any effective role matches allowed roles
		for _, userRole := range effectiveRoles {
			for _, allowedRole := range roles {
				if userRole == allowedRole {
					return next(ctx)
				}
			}
		}

		return nil, fmt.Errorf("unauthorized: user roles %v do not match required roles %v", effectiveRoles, roles)
	}
}

// RequireRole checks if the user has one of the required roles (legacy token-based version)
// Kept for backward compatibility with existing unit tests.
func RequireRole(ctx context.Context, obj any, next graphql.Resolver, roles []string) (any, error) {
	// Extract user roles from context (set by JWT middleware)
	userRoles, ok := ctx.Value(middleware.UserRolesKey).([]string)
	if !ok || len(userRoles) == 0 {
		return nil, fmt.Errorf("unauthorized: no roles found in context")
	}

	// Check if any of user's roles matches any of the allowed roles
	for _, userRole := range userRoles {
		for _, allowedRole := range roles {
			if userRole == allowedRole {
				// Role matches, proceed with the resolver
				return next(ctx)
			}
		}
	}

	// No matching role found
	return nil, fmt.Errorf("unauthorized: user roles %v do not match required roles %v", userRoles, roles)
}
