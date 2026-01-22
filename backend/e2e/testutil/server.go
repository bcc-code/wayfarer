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
	LanguageService    *services.LanguageService
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
			RequireRole: directives.NewRequireRole(cfg.Loaders.RolesByUserLoader),
		},
	}

	gqlHandler := handler.New(api.NewExecutableSchema(gqlConfig))
	gqlHandler.AddTransport(transport.POST{})

	// JWT config for testing
	jwtConfig := config.JWTConfig{
		Secret: TestJWTSecret,
		Issuer: TestJWTIssuer,
	}

	// GraphQL endpoint with language and JWT middleware
	router.POST("/graphql",
		middleware.LanguageExtractor(),
		middleware.JWTAuth(jwtConfig),
		graphqlHandlerWithLanguage(gqlHandler, cfg.LanguageService),
	)

	return router
}

// graphqlHandlerWithLanguage wraps a GraphQL handler for use with Gin, including language sync
func graphqlHandlerWithLanguage(h *handler.Server, languageService *services.LanguageService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Transfer Gin context values to request context for GraphQL resolvers
		ctx := c.Request.Context()

		// Transfer user_id if present
		var userID string
		if uid, exists := c.Get("user_id"); exists {
			userID = uid.(string)
			ctx = context.WithValue(ctx, middleware.UserIDKey, uid)
		}

		// Transfer user_roles if present
		if userRoles, exists := c.Get("user_roles"); exists {
			ctx = context.WithValue(ctx, middleware.UserRolesKey, userRoles)
		}

		// Transfer language if present
		var requestedLang string
		if language, exists := c.Get("language"); exists {
			requestedLang = language.(string)
			ctx = context.WithValue(ctx, middleware.LanguageKey, language)
		}

		// Sync language preference synchronously for testing (not fire-and-forget)
		if userID != "" && requestedLang != "" && languageService != nil {
			languageService.SyncUserLanguage(ctx, userID, requestedLang)
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
	languageService := services.NewLanguageService(dbMgr.DB.Queries, testCache, dataLoaders.UserByIDLoader, logger)

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
		LanguageService:    languageService,
		Loaders:            dataLoaders,
	})

	cleanup := func() {
		settingsService.Stop()
		testCache.Close()
	}

	return router, cleanup, nil
}

// SetupTestServerWithCache creates a complete test environment and returns the cache for testing
// Returns the router, cache, cleanup function, and error
func SetupTestServerWithCache(ctx context.Context, dbMgr *TestDBManager) (*gin.Engine, *cache.CacheWithRegistry, func(), error) {
	// Create cache
	testCache, err := NewTestCache()
	if err != nil {
		return nil, nil, nil, err
	}

	// Create loaders
	dataLoaders := loaders.NewLoaders(dbMgr.DB, testCache)

	// Create logger for test (discard output)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))

	// Create services
	roleService := services.NewRoleService(dbMgr.DB.Queries, testCache)
	leaderboardService := services.NewLeaderboardService(dbMgr.DB.Queries, testCache.Cache, dataLoaders)
	languageService := services.NewLanguageService(dbMgr.DB.Queries, testCache, dataLoaders.UserByIDLoader, logger)

	// Create settings service
	settingsService, err := services.NewSettingsService(ctx, dbMgr.DB.Queries, logger)
	if err != nil {
		testCache.Close()
		return nil, nil, nil, err
	}

	// Create router
	router := NewTestRouter(TestServerConfig{
		DB:                 dbMgr.DB,
		Cache:              testCache,
		RoleService:        roleService,
		LeaderboardService: leaderboardService,
		SettingsService:    settingsService,
		LanguageService:    languageService,
		Loaders:            dataLoaders,
	})

	cleanup := func() {
		settingsService.Stop()
		testCache.Close()
	}

	return router, testCache, cleanup, nil
}

// NewTestContentAchievementService creates a ContentAchievementService for testing
// without push notifications or webhook dispatching
func NewTestContentAchievementService(dbMgr *TestDBManager) (*services.ContentAchievementService, func(), error) {
	testCache, err := NewTestCache()
	if err != nil {
		return nil, nil, err
	}

	dataLoaders := loaders.NewLoaders(dbMgr.DB, testCache)

	service := &services.ContentAchievementService{
		DB:             dbMgr.DB,
		Cache:          testCache,
		PushService:    nil, // No push notifications in tests
		Loaders:        dataLoaders,
		WebhookService: nil, // No webhook dispatch in tests
	}

	cleanup := func() {
		testCache.Close()
	}

	return service, cleanup, nil
}
