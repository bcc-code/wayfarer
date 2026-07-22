package ladder_to_heaven

import (
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/plugins"
	"github.com/gin-gonic/gin"
)

// Config holds the configuration for the Ladder to Heaven plugin.
type Config struct {
	// AchievementID is the ID of the achievement to check for deadline bonus points.
	// If empty, the plugin is disabled.
	AchievementID string
	// SecretKey is used to verify webhook signatures.
	SecretKey string
	// TeamRenameChallengeID is the challenge to mark as completed when a team renames.
	// Awards 300 points to each team member (once per team).
	TeamRenameChallengeID string
	// CryptexSecretKey is used to sign Cryptex JWT tokens.
	// If empty, the cryptex admin URL endpoint is disabled.
	CryptexSecretKey string
	// CryptexBaseURL is the base URL for Cryptex admin login.
	// Example: https://cryptex.example.com
	CryptexBaseURL string
}

// LadderToHeavenPlugin implements the plugins.Plugin interface.
type LadderToHeavenPlugin struct {
	config Config
}

// NewPlugin creates a new Ladder to Heaven plugin instance.
func NewPlugin(cfg Config) *LadderToHeavenPlugin {
	return &LadderToHeavenPlugin{
		config: cfg,
	}
}

// Name returns the plugin's display name.
func (p *LadderToHeavenPlugin) Name() string {
	return "Ladder to Heaven"
}

// Enabled returns true if the plugin should be registered.
func (p *LadderToHeavenPlugin) Enabled() bool {
	return true
}

// Register sets up the plugin's routes.
func (p *LadderToHeavenPlugin) Register(router gin.IRouter, deps plugins.Dependencies, apiKeyAuth gin.HandlerFunc) error {
	// The webhook endpoints mutate state (award points, finalize betting), so they
	// must fail closed: only register them when a signing secret is configured, and
	// require a valid HMAC signature on every request. Without a secret the routes
	// are not registered at all (requests get 404).
	if p.config.SecretKey != "" {
		hmacAuth := middleware.WebhookHMACAuth(p.config.SecretKey)

		contentHandler := &contentEventHandler{
			db:            deps.DB,
			cache:         deps.Cache,
			achievementID: p.config.AchievementID,
			firebase:      deps.Firebase,
		}
		router.POST("/plugins/ladder-to-heaven/content-event", hmacAuth, contentHandler.handle)

		teamRenameHandler := &teamNameChangedHandler{
			db:          deps.DB,
			cache:       deps.Cache,
			challengeID: p.config.TeamRenameChallengeID,
			firebase:    deps.Firebase,
		}
		router.POST("/plugins/ladder-to-heaven/team-name-changed", hmacAuth, teamRenameHandler.handle)

		// Quiz finalized handler for ordering question betting
		quizFinalizedHandler := &quizFinalizedHandler{
			db:          deps.DB,
			cache:       deps.Cache,
			loaders:     deps.Loaders,
			pushService: deps.PushService,
			firebase:    deps.Firebase,
		}
		router.POST("/plugins/ladder-to-heaven/quiz-finalized", hmacAuth, quizFinalizedHandler.handle)
	} else {
		slog.Warn("ladder_to_heaven: webhook endpoints disabled, PLUGIN_LADDER_TO_HEAVEN_SECRET_KEY not set")
	}

	// Cryptex admin URL endpoint (requires JWT authentication)
	cryptexHandler := &cryptexAdminURLHandler{
		db:              deps.DB,
		settingsService: deps.SettingsService,
		secretKey:       p.config.CryptexSecretKey,
		baseURL:         p.config.CryptexBaseURL,
		jwtConfig:       deps.JWTConfig,
	}

	router.GET("/plugins/ladder-to-heaven/cryptex-admin-url", middleware.JWTAuth(deps.JWTConfig), cryptexHandler.handle)

	// Superteam distribution endpoints (requires JWT authentication)
	distHandler := &superteamDistributionHandler{
		db:        deps.DB,
		cache:     deps.Cache,
		jwtConfig: deps.JWTConfig,
	}

	router.GET("/plugins/ladder-to-heaven/preview-superteams", middleware.JWTAuth(deps.JWTConfig), distHandler.preview)
	router.POST("/plugins/ladder-to-heaven/distribute-superteams", middleware.JWTAuth(deps.JWTConfig), distHandler.handle)

	return nil
}
