package teams

import (
	"github.com/bcc-media/wayfarer/internal/plugins"
	"github.com/gin-gonic/gin"
)

// Config holds the configuration for the teams plugin.
type Config struct {
	// SecretKey is used to verify webhook signatures.
	SecretKey string
	// RenameChallengeID is the challenge to mark as completed when a team renames.
	// Awards 300 points to each team member (once per team).
	RenameChallengeID string
}

// TeamsPlugin implements the plugins.Plugin interface.
type TeamsPlugin struct {
	config Config
}

// NewPlugin creates a new teams plugin instance.
func NewPlugin(cfg Config) *TeamsPlugin {
	return &TeamsPlugin{
		config: cfg,
	}
}

// Name returns the plugin's display name.
func (p *TeamsPlugin) Name() string {
	return "Teams"
}

// Enabled returns true if the plugin should be registered.
func (p *TeamsPlugin) Enabled() bool {
	return true
}

// Register sets up the plugin's routes.
func (p *TeamsPlugin) Register(router gin.IRouter, deps plugins.Dependencies, apiKeyAuth gin.HandlerFunc) error {
	teamRenameHandler := &teamNameChangedHandler{
		db:          deps.DB,
		cache:       deps.Cache,
		challengeID: p.config.RenameChallengeID,
		secretKey:   p.config.SecretKey,
		firebase:    deps.Firebase,
	}

	router.POST("/plugins/teams/team-name-changed", teamRenameHandler.handle)
	// Old path, kept so webhooks configured before the move keep working
	router.POST("/plugins/ladder-to-heaven/team-name-changed", teamRenameHandler.handle)

	return nil
}
