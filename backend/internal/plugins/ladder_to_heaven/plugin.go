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
	// TeamRenameChallengeID is the challenge to mark as completed when a team renames.
	// Awards 300 points to each team member (once per team).
	TeamRenameChallengeID string
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
	contentHandler := &contentEventHandler{
		db:            deps.DB,
		cache:         deps.Cache,
		achievementID: p.config.AchievementID,
		secretKey:     p.config.SecretKey,
	}

	router.POST("/plugins/ladder-to-heaven/content-event", contentHandler.handle)

	teamRenameHandler := &teamNameChangedHandler{
		db:          deps.DB,
		cache:       deps.Cache,
		challengeID: p.config.TeamRenameChallengeID,
		secretKey:   p.config.SecretKey,
	}

	router.POST("/plugins/ladder-to-heaven/team-name-changed", teamRenameHandler.handle)

	return nil
}
