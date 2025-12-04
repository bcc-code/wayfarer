package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/bcc-media/wayfarer/internal/auth0"
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api"
	"github.com/bcc-media/wayfarer/internal/graph/directives"
	"github.com/bcc-media/wayfarer/internal/handlers"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/otel"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/ssf"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ravilushqa/otelgqlgen"
	"github.com/sony/gobreaker/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	// Initialize structured logger
	lgr := logger.New(cfg.Server.Environment, logger.ParseLevel(cfg.Log.Level))
	slog.SetDefault(lgr)

	// Initialize OpenTelemetry tracer
	ctx := context.Background()
	tracerProvider, err := otel.InitTracer(ctx, otel.Config{
		Enabled:          cfg.OTEL.Enabled,
		ServiceName:      cfg.OTEL.ServiceName,
		ServiceVersion:   cfg.OTEL.ServiceVersion,
		ExporterEndpoint: cfg.OTEL.ExporterEndpoint,
		ExporterInsecure: cfg.OTEL.ExporterInsecure,
		SamplingRatio:    cfg.OTEL.SamplingRatio,
	})
	if err != nil {
		slog.Error("Failed to initialize OpenTelemetry tracer", "error", err)
		os.Exit(1)
	}
	if tracerProvider != nil {
		defer func() {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := tracerProvider.Shutdown(shutdownCtx); err != nil {
				slog.Error("Failed to shutdown tracer provider", "error", err)
			}
		}()
	}

	// Connect to database
	db, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	slog.Info("Connected to database successfully")

	// Initialize JWKS for Brunstad TV JWT validation
	jwks, err := keyfunc.NewDefault([]string{cfg.JWT.BrunstadTVJWKSURL})
	if err != nil {
		slog.Error("Failed to initialize JWKS", "error", err)
		os.Exit(1)
	}
	slog.Info("JWKS initialized successfully", "url", cfg.JWT.BrunstadTVJWKSURL)

	// Initialize Auth0 client for Members API token management
	var membersClient *members.Client
	if cfg.Auth0.Domain != "" && cfg.Auth0.ClientID != "" && cfg.Members.Domain != "" {
		auth0Client := auth0.New(auth0.Config{
			Domain:       cfg.Auth0.Domain,
			ClientID:     cfg.Auth0.ClientID,
			ClientSecret: cfg.Auth0.ClientSecret,
		})

		// Create circuit breaker for Members API
		membersBreaker := gobreaker.NewCircuitBreaker[[]byte](gobreaker.Settings{
			Name:    "members-api",
			Timeout: 2 * time.Second,
		})

		// Initialize Members API client
		membersClient = members.New(
			members.Config{Domain: cfg.Members.Domain},
			auth0Client,
			membersBreaker,
		)
		slog.Info("Members API client initialized", "domain", cfg.Members.Domain)
	} else {
		slog.Warn("Members API client not initialized - missing configuration")
	}

	// Initialize SSF client and sync service
	var ssfSyncService *ssf.SyncService
	if cfg.SSF.APIKey != "" {
		ssfClient := ssf.New(ssf.Config{
			BaseURL:   cfg.SSF.BaseURL,
			APIKey:    cfg.SSF.APIKey,
			DebugMode: cfg.SSF.DebugMode,
			Timeout:   cfg.SSF.Timeout,
		}, lgr)
		slog.Info("SSF API client initialized",
			"base_url", cfg.SSF.BaseURL,
			"debug_mode", cfg.SSF.DebugMode,
		)

		ssfSyncService = ssf.NewSyncService(ssfClient, db.Queries, lgr)
	} else {
		slog.Warn("SSF API client not initialized - missing API key")
	}

	// Initialize cache with default configuration
	cacheInstance, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	if err != nil {
		slog.Error("Failed to initialize cache", "error", err)
		os.Exit(1)
	}
	defer cacheInstance.Close()
	slog.Info("Cache initialized", "default_ttl", "15m", "max_cost", "100MB")

	// Initialize DataLoaders (shared globally across all requests)
	// Dataloaders handle request batching while Ristretto cache handles data caching
	dataLoaders := loaders.NewLoaders(db, cacheInstance)
	slog.Info("DataLoaders initialized with Ristretto cache integration")

	// Initialize RoleService
	roleService := services.NewRoleService(db.Queries, cacheInstance)
	slog.Info("RoleService initialized with caching")

	// Initialize LeaderboardService
	leaderboardService := services.NewLeaderboardService(db.Queries, cacheInstance.Cache, dataLoaders)
	slog.Info("LeaderboardService initialized with caching and loaders")

	// Initialize GraphQL resolver
	apiResolver := &api.Resolver{
		DB:                 db,
		Loaders:            dataLoaders,
		Cache:              cacheInstance,
		RoleService:        roleService,
		LeaderboardService: leaderboardService,
	}

	apiHandler := handler.New(api.NewExecutableSchema(api.Config{
		Resolvers: apiResolver,
		Directives: api.DirectiveRoot{
			RequireRole: directives.RequireRole,
		},
	}))

	apiHandler.AddTransport(transport.Options{})
	apiHandler.AddTransport(transport.GET{})
	apiHandler.AddTransport(transport.POST{})
	apiHandler.Use(extension.Introspection{})
	apiHandler.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	apiHandler.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Add OpenTelemetry tracing for GraphQL operations
	if cfg.OTEL.Enabled {
		apiHandler.Use(otelgqlgen.Middleware())
		slog.Info("OpenTelemetry GraphQL instrumentation enabled")
	}

	// Set up Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add OpenTelemetry middleware for HTTP tracing
	if cfg.OTEL.Enabled {
		router.Use(otelgin.Middleware(cfg.OTEL.ServiceName))
		slog.Info("OpenTelemetry HTTP instrumentation enabled")
	}

	// Configure CORS to allow all headers
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*", "Authorization"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	// TODO: Actually check things, like DB connection
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Cache metrics endpoint
	router.GET("/metrics/cache", func(c *gin.Context) {
		metrics := cacheInstance.Metrics()
		hits := metrics.Hits()
		misses := metrics.Misses()
		total := hits + misses

		var hitRate float64
		if total > 0 {
			hitRate = float64(hits) / float64(total)
		}

		c.JSON(http.StatusOK, gin.H{
			"hits":          hits,
			"misses":        misses,
			"total":         total,
			"hit_rate":      hitRate,
			"hit_rate_pct":  hitRate * 100,
			"cost_added":    metrics.CostAdded(),
			"cost_evicted":  metrics.CostEvicted(),
			"keys_added":    metrics.KeysAdded(),
			"keys_updated":  metrics.KeysUpdated(),
			"keys_evicted":  metrics.KeysEvicted(),
			"sets_dropped":  metrics.SetsDropped(),
			"sets_rejected": metrics.SetsRejected(),
		})
	})

	// Profiling endpoints (use with caution in production)
	pprofGroup := router.Group("/debug/pprof")
	{
		pprofGroup.GET("/", gin.WrapF(pprof.Index))
		pprofGroup.GET("/cmdline", gin.WrapF(pprof.Cmdline))
		pprofGroup.GET("/profile", gin.WrapF(pprof.Profile))
		pprofGroup.POST("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/symbol", gin.WrapF(pprof.Symbol))
		pprofGroup.GET("/trace", gin.WrapF(pprof.Trace))
		pprofGroup.GET("/allocs", gin.WrapH(pprof.Handler("allocs")))
		pprofGroup.GET("/block", gin.WrapH(pprof.Handler("block")))
		pprofGroup.GET("/goroutine", gin.WrapH(pprof.Handler("goroutine")))
		pprofGroup.GET("/heap", gin.WrapH(pprof.Handler("heap")))
		pprofGroup.GET("/mutex", gin.WrapH(pprof.Handler("mutex")))
		pprofGroup.GET("/threadcreate", gin.WrapH(pprof.Handler("threadcreate")))
	}
	slog.Info("Profiling endpoints enabled at /debug/pprof")

	// Authentication callback endpoint (no JWT middleware)
	authHandler := &handlers.AuthHandler{
		DB:            db,
		Cfg:           cfg,
		JWKS:          jwks,
		MembersClient: membersClient,
		RoleService:   roleService,
	}
	router.GET("/callback", authHandler.Callback)

	// Webhook handler for external content events
	webhookHandler := &handlers.WebhookHandler{
		DB: db,
	}
	router.POST("/api/v1/content-events", middleware.APIKeyAuth(cfg.APIKey), webhookHandler.HandleContentEvent)

	// SSF sync endpoint (triggered by external cron/scheduler)
	if ssfSyncService != nil && cfg.SSF.SyncKey != "" {
		ssfHandler := &handlers.SSFHandler{
			SyncService: ssfSyncService,
			SyncKey:     cfg.SSF.SyncKey,
		}
		router.POST("/ssf/sync/:slug", ssfHandler.HandleSyncPlan)
		slog.Info("SSF sync endpoint registered at POST /ssf/sync/:slug")
	}

	// GraphQL API endpoint
	router.POST("/graphql", middleware.LanguageExtractor(), middleware.JWTAuth(cfg.JWT), graphqlHandler(apiHandler))
	if cfg.Server.Environment != "production" {
		router.GET("/graphql", gin.WrapH(playground.Handler("GraphQL API", "/graphql")))
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        router,
		ReadTimeout:    cfg.Server.ReadTimeout,
		WriteTimeout:   cfg.Server.WriteTimeout,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		slog.Info("Starting server",
			"address", addr,
			"environment", cfg.Server.Environment,
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	slog.Info("Server started successfully",
		"graphql_api", fmt.Sprintf("http://%s/graphql", addr),
	)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited")
}

// graphqlHandler wraps a GraphQL handler for use with Gin
func graphqlHandler(h *handler.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Transfer Gin context values to request context for GraphQL resolvers
		ctx := c.Request.Context()

		// Transfer user_id if present
		if userID, exists := c.Get("user_id"); exists {
			ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
		}

		// Transfer user_roles if present
		if userRoles, exists := c.Get("user_roles"); exists {
			ctx = context.WithValue(ctx, middleware.UserRolesKey, userRoles)
		}

		// Transfer language if present
		if language, exists := c.Get("language"); exists {
			ctx = context.WithValue(ctx, middleware.LanguageKey, language)
		}

		// Create new request with updated context
		r := c.Request.WithContext(ctx)
		h.ServeHTTP(c.Writer, r)
	}
}
