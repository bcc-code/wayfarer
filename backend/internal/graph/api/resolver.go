package api

import (
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/loaders"
	"github.com/bcc-media/wayfarer/internal/services"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	DB                 *database.DB
	Loaders            *loaders.Loaders
	Cache              *cache.CacheWithRegistry
	RoleService        *services.RoleService
	LeaderboardService *services.LeaderboardService
}
