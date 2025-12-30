package testutil

import (
	"context"
	"log/slog"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api"
	"github.com/bcc-media/wayfarer/internal/graph/directives"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/gin-gonic/gin"
)

// TestServerConfig holds configuration for the test server
type TestServerConfig struct {
	DB                 *database.DB
	Cache              *cache.CacheWithRegistry
	RoleService        *services.RoleService
	LeaderboardService *services.LeaderboardService
	SettingsService    *services.SettingsService
	Loaders            *loaders.Loaders
}

// NewTestCache creates a cache instance for testing
func NewTestCache() (*cache.CacheWithRegistry, error) {
	return cache.NewCacheWithRegistry(cache.Config{
		NumCounters: 10_000,
		MaxCost:     10_000_000, // 10MB
		BufferItems: 64,
		DefaultTTL:  cache.DefaultConfig().DefaultTTL,
	})
}

// NewTestRouter creates a Gin router configured for testing
func NewTestRouter(cfg TestServerConfig) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Create GraphQL handler
	resolver := &api.Resolver{
		DB:                 cfg.DB,
		Loaders:            cfg.Loaders,
		Cache:              cfg.Cache,
		RoleService:        cfg.RoleService,
		LeaderboardService: cfg.LeaderboardService,
		Settings:           cfg.SettingsService,
		InstanceID:         "test-instance",
	}

	gqlConfig := api.Config{
		Resolvers: resolver,
		Directives: api.DirectiveRoot{
			RequireRole: directives.RequireRole,
		},
	}

	gqlHandler := handler.New(api.NewExecutableSchema(gqlConfig))
	gqlHandler.AddTransport(transport.POST{})

	// JWT config for testing
	jwtConfig := config.JWTConfig{
		Secret: TestJWTSecret,
		Issuer: TestJWTIssuer,
	}

	// GraphQL endpoint with JWT middleware
	router.POST("/graphql", middleware.JWTAuth(jwtConfig), graphqlHandler(gqlHandler))

	return router
}

// graphqlHandler wraps a GraphQL handler for use with Gin
func graphqlHandler(h *handler.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Transfer Gin context values to request context for GraphQL resolvers
		ctx := c.Request.Context()

		// Transfer user_id if present
		if uid, exists := c.Get("user_id"); exists {
			ctx = context.WithValue(ctx, middleware.UserIDKey, uid)
		}

		// Transfer user_roles if present
		if userRoles, exists := c.Get("user_roles"); exists {
			ctx = context.WithValue(ctx, middleware.UserRolesKey, userRoles)
		}

		// Create new request with updated context
		r := c.Request.WithContext(ctx)
		h.ServeHTTP(c.Writer, r)
	}
}

// SetupTestServer creates a complete test environment
// Returns the router, cleanup function, and seeded data
func SetupTestServer(ctx context.Context, dbMgr *TestDBManager) (*gin.Engine, func(), error) {
	// Create cache
	testCache, err := NewTestCache()
	if err != nil {
		return nil, nil, err
	}

	// Create loaders
	dataLoaders := loaders.NewLoaders(dbMgr.DB, testCache)

	// Create logger for test (discard output)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create services
	roleService := services.NewRoleService(dbMgr.DB.Queries, testCache)
	leaderboardService := services.NewLeaderboardService(dbMgr.DB.Queries, testCache.Cache, dataLoaders)

	// Create settings service
	settingsService, err := services.NewSettingsService(ctx, dbMgr.DB.Queries, logger)
	if err != nil {
		testCache.Close()
		return nil, nil, err
	}

	// Create router
	router := NewTestRouter(TestServerConfig{
		DB:                 dbMgr.DB,
		Cache:              testCache,
		RoleService:        roleService,
		LeaderboardService: leaderboardService,
		SettingsService:    settingsService,
		Loaders:            dataLoaders,
	})

	cleanup := func() {
		settingsService.Stop()
		testCache.Close()
	}

	return router, cleanup, nil
}
