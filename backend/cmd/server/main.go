package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api"
	"github.com/bcc-media/wayfarer/internal/graph/directives"
	"github.com/bcc-media/wayfarer/internal/handlers"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker/v2"
	"github.com/vektah/gqlparser/v2/ast"
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

	// Connect to database
	ctx := context.Background()
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

	// Initialize DataLoaders (shared globally across all requests)
	dataLoaders := loaders.NewLoaders(db)
	slog.Info("DataLoaders initialized with global caching")

	// Initialize RoleService
	roleService := services.NewRoleService(db.Queries)
	slog.Info("RoleService initialized")

	// Initialize GraphQL resolver
	apiResolver := &api.Resolver{DB: db, Loaders: dataLoaders, RoleService: roleService}

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

	// Set up Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Configure CORS to allow all headers
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint
	// TODO: Actually check things, like DB connection
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Authentication callback endpoint (no JWT middleware)
	authHandler := &handlers.AuthHandler{
		DB:            db,
		Cfg:           cfg,
		JWKS:          jwks,
		MembersClient: membersClient,
		RoleService:   roleService,
	}
	router.GET("/callback", authHandler.Callback)

	// GraphQL API endpoint
	router.POST("/graphql", middleware.JWTAuth(cfg.JWT), graphqlHandler(apiHandler))
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
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
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

		// Create new request with updated context
		r := c.Request.WithContext(ctx)
		h.ServeHTTP(c.Writer, r)
	}
}
