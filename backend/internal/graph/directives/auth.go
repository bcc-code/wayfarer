package directives

import (
	"context"
	"fmt"

	"github.com/99designs/gqlgen/graphql"
	"github.com/bcc-media/wayfarer/internal/middleware"
)

// RequireRole checks if the user has one of the required roles
func RequireRole(ctx context.Context, obj any, next graphql.Resolver, roles []string) (any, error) {
	// Extract user role from context (set by JWT middleware)
	userRole, ok := ctx.Value(middleware.UserRoleKey).(string)
	if !ok || userRole == "" {
		return nil, fmt.Errorf("unauthorized: no role found in context")
	}

	// Check if user's role matches any of the allowed roles
	for _, allowedRole := range roles {
		if userRole == allowedRole {
			// Role matches, proceed with the resolver
			return next(ctx)
		}
	}

	// No matching role found
	return nil, fmt.Errorf("unauthorized: role '%s' is not allowed (required: %v)", userRole, roles)
}
