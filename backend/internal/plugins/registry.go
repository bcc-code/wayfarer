package plugins

import (
	"log/slog"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/firebase"
	"github.com/bcc-media/wayfarer/internal/services"
	"github.com/gin-gonic/gin"
)

// Dependencies contains all dependencies that plugins may need.
// Not all plugins will use all dependencies.
type Dependencies struct {
	DB              *database.DB
	Cache           *cache.CacheWithRegistry
	SettingsService *services.SettingsService
	JWTConfig       config.JWTConfig
	Firebase        *firebase.Service
}

// Plugin defines the interface that all plugins must implement.
type Plugin interface {
	// Name returns the plugin's display name for logging.
	Name() string

	// Enabled returns true if the plugin should be registered.
	// Plugins can check their configuration here.
	Enabled() bool

	// Register sets up the plugin's routes and returns any error.
	// This is only called if Enabled() returns true.
	Register(router gin.IRouter, deps Dependencies, apiKeyAuth gin.HandlerFunc) error
}

// RegisterPlugin registers a plugin with the router if it's enabled.
// Returns true if the plugin was registered, false if it was skipped.
func RegisterPlugin(router gin.IRouter, deps Dependencies, apiKeyAuth gin.HandlerFunc, plugin Plugin) bool {
	if !plugin.Enabled() {
		slog.Info("plugin disabled, skipping registration",
			"plugin", plugin.Name(),
		)
		return false
	}

	if err := plugin.Register(router, deps, apiKeyAuth); err != nil {
		slog.Error("failed to register plugin",
			"plugin", plugin.Name(),
			"error", err,
		)
		return false
	}

	slog.Info("plugin registered",
		"plugin", plugin.Name(),
	)
	return true
}
