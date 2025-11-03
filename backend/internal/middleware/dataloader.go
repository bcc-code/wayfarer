package middleware

import (
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/gin-gonic/gin"
)

// DataLoader creates middleware that injects DataLoaders into the request context
// DataLoaders batch database queries to avoid N+1 query problems
func DataLoader(db *database.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create new loaders for each request
		dataLoaders := loaders.NewLoaders(db)

		// Inject into context
		ctx := loaders.Inject(c.Request.Context(), dataLoaders)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
