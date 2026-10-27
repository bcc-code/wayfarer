package ssf

import (
	"github.com/bcc-media/wayfarer/internal/plugins"
	"github.com/gin-gonic/gin"
)

// Config holds the configuration for the SSF plugin.
type Config struct {
	// AchievementID is the ID of the achievement to check for deadline bonus points.
	// If empty, the content-event webhook is disabled.
	AchievementID string
	// SecretKey is used to verify webhook signatures.
	SecretKey string
}

// SSFPlugin implements the plugins.Plugin interface.
type SSFPlugin struct {
	config Config
}

// NewPlugin creates a new SSF plugin instance.
func NewPlugin(cfg Config) *SSFPlugin {
	return &SSFPlugin{
		config: cfg,
	}
}

// Name returns the plugin's display name.
func (p *SSFPlugin) Name() string {
	return "SSF"
}

// Enabled returns true if the plugin should be registered.
func (p *SSFPlugin) Enabled() bool {
	return true
}

// Register sets up the plugin's routes.
func (p *SSFPlugin) Register(router gin.IRouter, deps plugins.Dependencies, apiKeyAuth gin.HandlerFunc) error {
	contentHandler := &contentEventHandler{
		db:            deps.DB,
		cache:         deps.Cache,
		achievementID: p.config.AchievementID,
		secretKey:     p.config.SecretKey,
		firebase:      deps.Firebase,
	}

	router.POST("/plugins/ssf/content-event", contentHandler.handle)
	// Old path, kept so webhooks configured before the move keep working
	router.POST("/plugins/ladder-to-heaven/content-event", contentHandler.handle)

	return nil
}
