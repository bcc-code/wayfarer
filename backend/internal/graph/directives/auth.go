package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

// RequireRole checks if the user has one of the required roles
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
