package excalibur

import (
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/plugins"
	"github.com/gin-gonic/gin"
)

// Config holds the configuration for the Excalibur plugin.
type Config struct {
	// SecretKey is used to sign Excalibur JWT tokens.
	// If empty, the user URL endpoint responds with "feature disabled".
	SecretKey string
	// BaseURL is the base URL for Excalibur user login.
	// Example: https://dev.excalibur.bcc.media
	BaseURL string
}

// ExcaliburPlugin implements the plugins.Plugin interface.
type ExcaliburPlugin struct {
	config Config
}

// NewPlugin creates a new Excalibur plugin instance.
func NewPlugin(cfg Config) *ExcaliburPlugin {
	return &ExcaliburPlugin{
		config: cfg,
	}
}

// Name returns the plugin's display name.
func (p *ExcaliburPlugin) Name() string {
	return "Excalibur"
}

// Enabled returns true if the plugin should be registered.
func (p *ExcaliburPlugin) Enabled() bool {
	return true
}

// Register sets up the plugin's routes.
func (p *ExcaliburPlugin) Register(router gin.IRouter, deps plugins.Dependencies, apiKeyAuth gin.HandlerFunc) error {
	// User login URL endpoint (requires JWT authentication)
	handler := &excaliburUserURLHandler{
		db:              deps.DB,
		settingsService: deps.SettingsService,
		secretKey:       p.config.SecretKey,
		baseURL:         p.config.BaseURL,
		jwtConfig:       deps.JWTConfig,
	}

	router.GET("/plugins/excalibur/user-url", middleware.JWTAuth(deps.JWTConfig), handler.handle)

	return nil
}
