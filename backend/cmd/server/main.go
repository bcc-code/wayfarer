package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/MicahParks/jwkset"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/bcc-media/wayfarer/internal/auth0"
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/firebase"
	"github.com/bcc-media/wayfarer/internal/graph/api"
	"github.com/bcc-media/wayfarer/internal/graph/directives"
	"github.com/bcc-media/wayfarer/internal/handlers"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/loadtestauth"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/members"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/otel"
	"github.com/bcc-media/wayfarer/internal/plugins"
	"github.com/bcc-media/wayfarer/internal/plugins/ladder_to_heaven"
	"github.com/bcc-media/wayfarer/internal/pubsub"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/bcc-media/wayfarer/internal/services/bulk"
	"github.com/bcc-media/wayfarer/internal/services/email"
	"github.com/bcc-media/wayfarer/internal/services/push"
	"github.com/bcc-media/wayfarer/internal/services/webhooks"
	"github.com/bcc-media/wayfarer/internal/ssf"
	"github.com/bcc-media/wayfarer/translations"
	"github.com/bcc-media/wayfarer/translations/phrase"
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

	// Run database migrations
	if err := database.Migrate(ctx, cfg.Database.URL); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		os.Exit(1)
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
	// Use background context to avoid request context deadline issues
	jwks, err := keyfunc.NewDefaultCtx(ctx, []string{cfg.JWT.BrunstadTVJWKSURL})
	if err != nil {
		slog.Error("Failed to initialize Brunstad TV JWKS", "error", err)
		os.Exit(1)
	}
	slog.Info("Brunstad TV JWKS initialized successfully", "url", cfg.JWT.BrunstadTVJWKSURL)

	// Initialize JWKS for Auth0 (login.bcc.no) JWT validation
	// Use custom storage with SkipAll to handle X5T validation issues in Auth0's JWKS
	// Core security (RSA signature verification, expiration) is still enforced by JWT parsing
	auth0Options := jwkset.HTTPClientStorageOptions{
		Ctx: ctx,
		ValidateOptions: jwkset.JWKValidateOptions{
			SkipAll: true, // Skip JWK validation that fails with Auth0's X5T mismatch
		},
	}
	var loadtestJWKS []byte
	if cfg.JWT.Auth0LoadtestKey != "" {
		// Load-test mode: derive the JWKS from the shared load-test key and
		// serve it at /jwks.json ourselves, so the boot-time fetch of
		// Auth0JWKSURL may target this very process. Tolerate the failed
		// first fetch and refresh in the background; the keyfunc client also
		// re-fetches on an unknown kid.
		key, err := loadtestauth.ParsePrivateKey(cfg.JWT.Auth0LoadtestKey)
		if err != nil {
			slog.Error("Failed to parse AUTH0_LOADTEST_PRIVATE_KEY", "error", err)
			os.Exit(1)
		}
		loadtestJWKS, err = loadtestauth.BuildJWKS(&key.PublicKey)
		if err != nil {
			slog.Error("Failed to build load-test JWKS", "error", err)
			os.Exit(1)
		}
		auth0Options.NoErrorReturnFirstHTTPReq = true
		auth0Options.RefreshInterval = 30 * time.Second
	}
	auth0Storage, err := jwkset.NewStorageFromHTTP(cfg.JWT.Auth0JWKSURL, auth0Options)
	if err != nil {
		slog.Error("Failed to create Auth0 JWKS storage", "error", err)
		os.Exit(1)
	}
	auth0JWKS, err := keyfunc.New(keyfunc.Options{
		Ctx:     ctx,
		Storage: auth0Storage,
	})
	if err != nil {
		slog.Error("Failed to initialize Auth0 JWKS", "error", err)
		os.Exit(1)
	}
	slog.Info("Auth0 JWKS initialized successfully", "url", cfg.JWT.Auth0JWKSURL)

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
	var ssfClient *ssf.Client
	var ssfSyncService *ssf.SyncService
	if cfg.SSF.APIKey != "" {
		ssfClient = ssf.New(ssf.Config{
			BaseURL:   cfg.SSF.BaseURL,
			APIKey:    cfg.SSF.APIKey,
			DebugMode: cfg.SSF.DebugMode,
			Timeout:   cfg.SSF.Timeout,
		}, lgr)
		slog.Info("SSF API client initialized",
			"base_url", cfg.SSF.BaseURL,
			"debug_mode", cfg.SSF.DebugMode,
		)

		ssfSyncService = ssf.NewSyncService(ssfClient, db.Queries, db.Pool, lgr)
	} else {
		slog.Warn("SSF API client not initialized - missing API key")
	}

	// Initialize translations service with Phrase provider
	var translationsService *translations.Service
	if cfg.Phrase.Enabled && cfg.Phrase.ProjectUID != "" {
		phraseClient := phrase.NewClient(
			db.Queries,
			phrase.Config{
				BaseURL:     cfg.Phrase.BaseURL,
				Username:    cfg.Phrase.Username,
				Password:    cfg.Phrase.Password,
				ProjectUID:  cfg.Phrase.ProjectUID,
				CallbackURL: cfg.Phrase.CallbackURL,
				UserUID:     cfg.Phrase.UserUID,
				Debug:       cfg.Phrase.Debug,
			},
			cfg.Phrase.Languages,
		)
		phraseClient.SetDebug(cfg.Phrase.Debug)
		slog.Info("Phrase translation client initialized",
			"base_url", cfg.Phrase.BaseURL,
			"project_uid", cfg.Phrase.ProjectUID,
			"languages", cfg.Phrase.Languages,
		)

		translationsService = translations.NewService(db.Queries, phraseClient)
		slog.Info("Translation service initialized")
	} else {
		slog.Warn("Translation service not initialized - Phrase not enabled or missing configuration")
	}

	// Initialize cache with default configuration
	cacheInstance, err := cache.NewCacheWithRegistry(cache.DefaultConfig())
	if err != nil {
		slog.Error("Failed to initialize cache", "error", err)
		os.Exit(1)
	}
	defer cacheInstance.Close()
	slog.Info("Cache initialized", "default_ttl", "15m", "max_cost", "100MB")

	// Initialize cache sync for cross-instance invalidation via PostgreSQL NOTIFY/LISTEN
	cacheSync := cache.NewCacheSync(cacheInstance, cfg.Database.URL)
	if err := cacheSync.Start(ctx); err != nil {
		slog.Warn("Failed to start cache sync, continuing without cross-instance invalidation", "error", err)
	} else {
		cacheInstance.SetSync(cacheSync, db.Pool)
		defer cacheSync.Stop()
	}

	// Initialize DataLoaders (shared globally across all requests)
	// Dataloaders handle request batching while Ristretto cache handles data caching
	dataLoaders := loaders.NewLoaders(db, cacheInstance)
	slog.Info("DataLoaders initialized with Ristretto cache integration")

	// Initialize RoleService
	roleService := services.NewRoleService(db.Queries, cacheInstance)
	slog.Info("RoleService initialized with caching")

	// Initialize LeaderboardService
	leaderboardService := services.NewLeaderboardService(db.Queries, cacheInstance, dataLoaders)
	slog.Info("LeaderboardService initialized with caching and loaders")

	// Initialize SettingsService
	settingsService, err := services.NewSettingsService(ctx, db.Queries, lgr)
	if err != nil {
		slog.Error("Failed to initialize SettingsService", "error", err)
		os.Exit(1)
	}
	defer settingsService.Stop()
	slog.Info("SettingsService initialized with background refresh")

	// Initialize S3 upload service
	s3Service, err := services.NewS3Service(cfg.S3)
	if err != nil {
		slog.Error("Failed to initialize S3 service", "error", err)
		os.Exit(1)
	}
	slog.Info("S3 service initialized", "bucket", cfg.S3.Bucket, "region", cfg.S3.Region)

	// Initialize push notification service
	var pushService *push.Service
	if cfg.VAPID.PublicKey != "" && cfg.VAPID.PrivateKey != "" {
		pushService = push.NewService(db.Queries, cfg.VAPID, lgr)
		slog.Info("Push notification service initialized")
	} else {
		slog.Warn("Push notification service not initialized - VAPID keys not configured")
	}

	// Initialize LanguageService
	languageService := services.NewLanguageService(db.Queries, cacheInstance, dataLoaders.UserByIDLoader, lgr)
	slog.Info("LanguageService initialized")

	// Initialize WebhookService
	webhookService := webhooks.NewService(db.Queries)
	slog.Info("WebhookService initialized")

	// Initialize Firebase service (optional)
	var firebaseService *firebase.Service
	if cfg.Firebase.ServiceAccountJSON != "" {
		var err error
		firebaseService, err = firebase.New(ctx, cfg.Firebase)
		if err != nil {
			slog.Error("Failed to initialize Firebase service", "error", err)
			os.Exit(1)
		}
		slog.Info("Firebase service initialized")
	} else {
		slog.Warn("Firebase service not initialized - missing configuration")
	}

	// Keep a Firebase custom token cached for every user in the current project.
	// Minting is a local RSA-2048 signature (~1ms of CPU); a spike of cold users
	// hitting firebaseToken at once turns that into thousands of signatures per
	// second — a CPU profile of the 10k-user quiz spike attributed ~25% of all
	// server CPU to it. Rotating in the background costs a handful of signatures
	// per second and makes the request path a cache hit.
	var tokenWarmer *services.FirebaseTokenWarmer
	if firebaseService != nil {
		tokenWarmer = services.NewFirebaseTokenWarmer(firebaseService, db.Queries, cacheInstance, settingsService)
		if tokenWarmer != nil {
			tokenWarmer.Start(ctx)
			defer tokenWarmer.Stop()
			slog.Info("Firebase token warmer started")
		}
	}

	// Initialize Email service for feedback forwarding
	var emailService *email.Service
	if cfg.Resend.APIKey != "" {
		emailService = email.NewService(cfg.Resend.APIKey, cfg.Resend.AdminBaseURL, cfg.Resend.SSFTicketEmail)
		slog.Info("Email service initialized")
	} else {
		slog.Warn("Email service not configured - RESEND_API_KEY not set")
	}

	// Initialize Pub/Sub publisher for async bulk operations
	var pubsubPublisher *pubsub.Publisher
	if cfg.PubSub.Enabled {
		var err error
		pubsubPublisher, err = pubsub.NewPublisher(ctx, pubsub.PublisherConfig{
			Enabled:   cfg.PubSub.Enabled,
			ProjectID: cfg.PubSub.ProjectID,
			TopicID:   cfg.PubSub.TopicID,
		}, lgr)
		if err != nil {
			slog.Error("Failed to initialize Pub/Sub publisher", "error", err)
			os.Exit(1)
		}
		defer pubsubPublisher.Close()
		slog.Info("Pub/Sub publisher initialized",
			"project_id", cfg.PubSub.ProjectID,
			"topic_id", cfg.PubSub.TopicID,
		)
	} else {
		slog.Warn("Pub/Sub publisher not initialized - PUBSUB_ENABLED is false")
	}

	// Content achievement service (shared between auth and webhook handlers)
	contentAchievementService := &services.ContentAchievementService{
		DB:             db,
		Cache:          cacheInstance,
		PushService:    pushService,
		Loaders:        dataLoaders,
		WebhookService: webhookService,
	}

	// Church resolver (shared between auth, maintenance, and sync)
	churchResolver := &services.ChurchResolver{
		DB:            db,
		MembersClient: membersClient,
	}

	// User sync service
	userSyncService := &services.UserSyncService{
		DB:                        db,
		Cache:                     cacheInstance,
		SSFClient:                 ssfClient,
		MembersClient:             membersClient,
		ChurchResolver:            churchResolver,
		ContentAchievementService: contentAchievementService,
	}

	// Initialize bulk operations service
	bulkService := bulk.NewService(
		db,
		cacheInstance,
		dataLoaders,
		pushService,
		firebaseService,
		pubsubPublisher,
		lgr,
		contentAchievementService,
	)
	slog.Info("Bulk operations service initialized")

	// Initialize GraphQL resolver
	apiResolver := &api.Resolver{
		DB:                 db,
		Loaders:            dataLoaders,
		Cache:              cacheInstance,
		RoleService:        roleService,
		LeaderboardService: leaderboardService,
		Settings:           settingsService,
		PushService:        pushService,
		WebhookService:     webhookService,
		FirebaseService:    firebaseService,
		EmailService:       emailService,
		UserSyncService:    userSyncService,
		BulkService:        bulkService,
		InstanceID:         cacheSync.InstanceID(),
	}

	apiHandler := handler.New(api.NewExecutableSchema(api.Config{
		Resolvers: apiResolver,
		Directives: api.DirectiveRoot{
			RequireRole: directives.NewRequireRole(dataLoaders.RolesByUserLoader),
		},
	}))

	apiHandler.AddTransport(transport.Options{})
	apiHandler.AddTransport(transport.GET{})
	apiHandler.AddTransport(transport.POST{})
	apiHandler.Use(extension.Introspection{})
	apiHandler.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})
	// Whole-response cache for queries whose result is identical for every
	// caller (currently the current-project bootstrap query). 30s TTL bounds
	// staleness after admin edits.
	apiHandler.Use(api.ResponseCache{Cache: cacheInstance, TTL: 30 * time.Second})
	apiHandler.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	// Add OpenTelemetry tracing for GraphQL operations
	if cfg.OTEL.Enabled {
		apiHandler.Use(otelgqlgen.Middleware())
		slog.Info("OpenTelemetry GraphQL instrumentation enabled")
	}

	// Set up Gin router. In production, skip gin's per-request access logger:
	// it writes one line per request to stdout synchronously (release mode
	// does not disable it), which at load-test rates is thousands of journald
	// writes per second.
	var router *gin.Engine
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
		router.Use(gin.Recovery())
	} else {
		router = gin.Default()
	}

	// Aggregated per-route request stats (served at /metrics/http) plus
	// error/slow-request logging — the production replacement for gin's
	// per-request access log.
	httpStats := middleware.NewHTTPStats()
	router.Use(httpStats.Middleware())
	if cfg.Server.HTTPStatsFile != "" {
		httpStats.StartDumper(ctx, cfg.Server.HTTPStatsFile, cfg.Server.HTTPStatsInterval)
		slog.Info("HTTP stats dumper enabled",
			"file", cfg.Server.HTTPStatsFile, "interval", cfg.Server.HTTPStatsInterval)
	}

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

	// Aggregated HTTP request stats (counts, status classes, latency
	// percentiles per route since boot). Loopback-only: the route inventory
	// and traffic volumes are operational data, not public API — read it on
	// the host (curl localhost) or over an SSH tunnel.
	router.GET("/metrics/http", func(c *gin.Context) {
		if ip := net.ParseIP(c.ClientIP()); ip == nil || !ip.IsLoopback() {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusOK, httpStats.Snapshot())
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

	// Authentication token endpoint (no JWT middleware)
	authHandler := &handlers.AuthHandler{
		DB:                        db,
		Cfg:                       cfg,
		JWKS:                      jwks,
		Auth0JWKS:                 auth0JWKS,
		MembersClient:             membersClient,
		RoleService:               roleService,
		ContentAchievementService: contentAchievementService,
		ChurchResolver:            churchResolver,
		UserSyncService:           userSyncService,
	}
	router.GET("/token", authHandler.Callback)

	// Load-test only: serve the simulated-Auth0 JWKS derived from
	// AUTH0_LOADTEST_PRIVATE_KEY, so AUTH0_JWKS_URL can point back at this
	// server (no sidecar or file needed). See internal/loadtestauth.
	if loadtestJWKS != nil {
		slog.Info("Serving load-test JWKS", "route", "/jwks.json")
		router.GET("/jwks.json", func(c *gin.Context) {
			c.Data(http.StatusOK, "application/json", loadtestJWKS)
		})
	}

	// Webhook handler for external content events
	webhookHandler := &handlers.WebhookHandler{
		DB:                        db,
		WebhookService:            webhookService,
		ContentAchievementService: contentAchievementService,
	}
	router.POST("/api/v1/content-events", middleware.APIKeyAuth(cfg.APIKey), webhookHandler.HandleContentEvent)

	// Consent webhook handler for external consent events
	consentWebhookHandler := &handlers.ConsentWebhookHandler{
		DB:              db,
		Cache:           cacheInstance,
		UserSyncService: userSyncService,
	}
	router.POST("/api/v1/consent-events", middleware.APIKeyAuth(cfg.APIKey), consentWebhookHandler.HandleConsentEvent)

	// Register plugins
	pluginDeps := plugins.Dependencies{
		DB:              db,
		Cache:           cacheInstance,
		Loaders:         dataLoaders,
		PushService:     pushService,
		SettingsService: settingsService,
		JWTConfig:       cfg.JWT,
		Firebase:        firebaseService,
	}
	apiKeyAuth := middleware.APIKeyAuth(cfg.APIKey)

	plugins.RegisterPlugin(router, pluginDeps, apiKeyAuth,
		ladder_to_heaven.NewPlugin(ladder_to_heaven.Config{
			AchievementID:         cfg.Plugin.LadderToHeavenAchievementID,
			SecretKey:             cfg.Plugin.LadderToHeavenSecretKey,
			TeamRenameChallengeID: cfg.Plugin.LadderToHeavenTeamRenameChallengeID,
			CryptexSecretKey:      cfg.Plugin.LadderToHeavenCryptexSecretKey,
			CryptexBaseURL:        cfg.Plugin.LadderToHeavenCryptexBaseURL,
			ExcaliburBaseURL:      cfg.Plugin.LadderToHeavenExcaliburBaseURL,
		}),
	)

	// Maintenance handler for syncing user data from Members API
	maintenanceHandler := &handlers.MaintenanceHandler{
		DB:                        db,
		Cache:                     cacheInstance,
		MembersClient:             membersClient,
		ChurchResolver:            churchResolver,
		AuthHandler:               authHandler,
		ContentAchievementService: contentAchievementService,
		SSFClient:                 ssfClient,
	}
	router.POST("/api/maintenance/sync-user-data", middleware.APIKeyAuth(cfg.APIKey), maintenanceHandler.SyncUserData)
	router.POST("/api/maintenance/sync-user/:user_id", middleware.APIKeyAuth(cfg.APIKey), maintenanceHandler.SyncSingleUser)
	router.POST("/api/maintenance/backfill-ssf-events", middleware.APIKeyAuth(cfg.APIKey), maintenanceHandler.BackfillSSFEvents)
	slog.Info("Maintenance endpoints registered",
		"batch_sync", "POST /api/maintenance/sync-user-data",
		"single_sync", "POST /api/maintenance/sync-user/:user_id",
		"backfill_ssf", "POST /api/maintenance/backfill-ssf-events",
	)

	// Quiz scheduler handler for timed session state transitions
	quizSchedulerHandler := &handlers.QuizSchedulerHandler{
		DB:              db,
		FirebaseService: firebaseService,
		WebhookService:  webhookService,
		Cache:           cacheInstance,
	}
	router.POST("/api/scheduler/quiz-session-transitions", middleware.APIKeyAuth(cfg.APIKey), quizSchedulerHandler.ProcessScheduledTransitions)
	slog.Info("Quiz scheduler endpoint registered",
		"endpoint", "POST /api/scheduler/quiz-session-transitions",
	)

	// Pub/Sub push handler for bulk operations (always registered - for receiving messages)
	// This is separate from the publisher initialization which requires PUBSUB_ENABLED=true
	bulkProcessor := bulk.NewProcessor(bulkService)
	pubsubProcessor := pubsub.NewProcessor(db, bulkProcessor, lgr)
	pubsubHandler := pubsub.NewHandler(pubsubProcessor, lgr)
	router.POST("/pubsub/bulk-operations", pubsubHandler.HandlePush)
	slog.Info("Pub/Sub bulk operations endpoint registered",
		"endpoint", "POST /pubsub/bulk-operations",
	)

	// File upload handler
	uploadHandler := &handlers.UploadHandler{
		DB:        db,
		S3Service: s3Service,
	}
	router.POST("/api/upload", middleware.JWTAuth(cfg.JWT), uploadHandler.HandleFileUpload)

	// SSF sync endpoint (triggered by external cron/scheduler)
	if ssfSyncService != nil && cfg.SSF.SyncKey != "" {
		ssfHandler := &handlers.SSFHandler{
			SyncService: ssfSyncService,
			SyncKey:     cfg.SSF.SyncKey,
		}
		router.POST("/ssf/sync/:slug", ssfHandler.HandleSyncPlan)
		slog.Info("SSF sync endpoint registered at POST /ssf/sync/:slug")
	}

	// Translation endpoints (export and webhook)
	if translationsService != nil && cfg.Phrase.ExportKey != "" {
		translationsHandler := handlers.NewTranslationsHandler(translationsService, cfg.Phrase.ExportKey)
		router.POST("/api/translations/webhook", translationsHandler.HandleWebhook)
		router.POST("/api/translations/export/all", translationsHandler.HandleExportAll)
		router.POST("/api/translations/export/:collection", translationsHandler.HandleExport)
		slog.Info("Translation endpoints registered",
			"webhook", "POST /api/translations/webhook",
			"export_all", "POST /api/translations/export/all",
			"export_collection", "POST /api/translations/export/:collection",
		)
	}

	// GraphQL API endpoint
	router.POST("/graphql", middleware.LanguageExtractor(), middleware.JWTAuth(cfg.JWT), graphqlHandler(apiHandler, languageService))
	if cfg.Server.Environment != "production" {
		router.GET("/graphql", gin.WrapH(playground.Handler("GraphQL API", "/graphql")))
	}

	// Serve frontend static files if configured
	if cfg.Server.StaticFilesPath != "" {
		staticPath := cfg.Server.StaticFilesPath
		indexPath := filepath.Join(staticPath, "index.html")

		if _, err := os.Stat(indexPath); err == nil {
			slog.Info("Serving static files", "path", staticPath)

			// Serve Nuxt build output directories
			router.Static("/_nuxt", filepath.Join(staticPath, "_nuxt"))
			router.Static("/_fonts", filepath.Join(staticPath, "_fonts"))
			router.Static("/images", filepath.Join(staticPath, "images"))

			// Serve root static files
			router.StaticFile("/favicon.ico", filepath.Join(staticPath, "favicon.ico"))
			router.StaticFile("/favicon.svg", filepath.Join(staticPath, "favicon.svg"))
			router.StaticFile("/manifest.webmanifest", filepath.Join(staticPath, "manifest.webmanifest"))
			router.StaticFile("/robots.txt", filepath.Join(staticPath, "robots.txt"))
			router.StaticFile("/apple-touch-icon-180x180.png", filepath.Join(staticPath, "apple-touch-icon-180x180.png"))
			router.StaticFile("/pwa-64x64.png", filepath.Join(staticPath, "pwa-64x64.png"))
			router.StaticFile("/pwa-192x192.png", filepath.Join(staticPath, "pwa-192x192.png"))
			router.StaticFile("/pwa-512x512.png", filepath.Join(staticPath, "pwa-512x512.png"))
			router.StaticFile("/maskable-icon-512x512.png", filepath.Join(staticPath, "maskable-icon-512x512.png"))
			router.StaticFile("/service-worker.js", filepath.Join(staticPath, "service-worker.js"))

			// SPA fallback: serve index.html for unmatched routes
			router.NoRoute(func(c *gin.Context) {
				// Only serve index.html for GET requests
				if c.Request.Method != http.MethodGet {
					c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
					return
				}

				// Skip API paths that weren't matched
				path := c.Request.URL.Path
				if strings.HasPrefix(path, "/api/") ||
					strings.HasPrefix(path, "/graphql") ||
					strings.HasPrefix(path, "/debug/") ||
					strings.HasPrefix(path, "/metrics/") ||
					strings.HasPrefix(path, "/ssf/") {
					c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
					return
				}

				c.File(indexPath)
			})
		} else {
			slog.Warn("Static files not found, frontend serving disabled", "path", staticPath)
		}
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
		tls := cfg.Server.TLSCertFile != "" && cfg.Server.TLSKeyFile != ""
		slog.Info("Starting server",
			"address", addr,
			"environment", cfg.Server.Environment,
			"tls", tls,
		)
		var err error
		if tls {
			// Native TLS termination — no fronting proxy.
			err = srv.ListenAndServeTLS(cfg.Server.TLSCertFile, cfg.Server.TLSKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
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
func graphqlHandler(h *handler.Server, languageService *services.LanguageService) gin.HandlerFunc {
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
		var userRoles []string
		if roles, exists := c.Get("user_roles"); exists {
			userRoles = roles.([]string)
			ctx = context.WithValue(ctx, middleware.UserRolesKey, roles)
		}

		// Transfer language if present
		var requestedLang string
		if language, exists := c.Get("language"); exists {
			requestedLang = language.(string)
			ctx = context.WithValue(ctx, middleware.LanguageKey, language)
		}

		// Transfer User-Agent header
		userAgent := c.GetHeader("User-Agent")
		ctx = context.WithValue(ctx, middleware.UserAgentKey, userAgent)

		// Sync language preference asynchronously (fire-and-forget)
		// Skip M2M users as they don't exist in the database
		isM2MUser := false
		for _, role := range userRoles {
			if role == "m2m" {
				isM2MUser = true
				break
			}
		}
		if userID != "" && requestedLang != "" && languageService != nil && !isM2MUser {
			go languageService.SyncUserLanguage(context.Background(), userID, requestedLang)
		}

		// Create new request with updated context
		r := c.Request.WithContext(ctx)
		h.ServeHTTP(c.Writer, r)
	}
}
