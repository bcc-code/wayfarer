package ladder_to_heaven

import (
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
	return p.config.AchievementID != ""
}

// Register sets up the plugin's routes.
func (p *LadderToHeavenPlugin) Register(router gin.IRouter, deps plugins.Dependencies, apiKeyAuth gin.HandlerFunc) error {
	handler := &contentEventHandler{
		db:            deps.DB,
		cache:         deps.Cache,
		achievementID: p.config.AchievementID,
		secretKey:     p.config.SecretKey,
	}

	router.POST("/plugins/ladder-to-heaven/content-event", handler.handle)

	return nil
}
