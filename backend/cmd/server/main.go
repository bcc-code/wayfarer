package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/admin"
	"github.com/bcc-media/wayfarer/internal/graph/m2m"
	"github.com/bcc-media/wayfarer/internal/graph/user"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
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

	// Initialize GraphQL resolvers
	userResolver := &user.Resolver{DB: db}
	adminResolver := &admin.Resolver{DB: db}
	m2mResolver := &m2m.Resolver{DB: db}

	// Create GraphQL handlers
	userHandler := handler.NewDefaultServer(user.NewExecutableSchema(user.Config{Resolvers: userResolver}))
	adminHandler := handler.NewDefaultServer(admin.NewExecutableSchema(admin.Config{Resolvers: adminResolver}))
	m2mHandler := handler.NewDefaultServer(m2m.NewExecutableSchema(m2m.Config{Resolvers: m2mResolver}))

	// Set up Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// User API endpoints
	router.POST("/graphql/user", middleware.DataLoader(db), middleware.JWTAuth(cfg.JWT), graphqlHandler(userHandler))
	if cfg.Server.Environment != "production" {
		router.GET("/graphql/user", gin.WrapH(playground.Handler("User API", "/graphql/user")))
	}

	// Admin API endpoints
	router.POST("/graphql/admin", middleware.JWTAuth(cfg.JWT), graphqlHandler(adminHandler))
	if cfg.Server.Environment != "production" {
		router.GET("/graphql/admin", gin.WrapH(playground.Handler("Admin API", "/graphql/admin")))
	}

	// M2M API endpoints
	router.POST("/graphql/m2m", middleware.JWTAuth(cfg.JWT), graphqlHandler(m2mHandler))
	if cfg.Server.Environment != "production" {
		router.GET("/graphql/m2m", gin.WrapH(playground.Handler("M2M API", "/graphql/m2m")))
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
		"user_api", fmt.Sprintf("http://%s/graphql/user", addr),
		"admin_api", fmt.Sprintf("http://%s/graphql/admin", addr),
		"m2m_api", fmt.Sprintf("http://%s/graphql/m2m", addr),
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
		h.ServeHTTP(c.Writer, c.Request)
	}
}
