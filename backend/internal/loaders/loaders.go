package loaders

import (
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// Loaders holds all dataloader instances for batching database queries
// These are shared globally across all requests and use built-in caching
type Loaders struct {
	UserByIDLoader       *dataloader.Loader[string, *model.User]
	ChurchLoader         *dataloader.Loader[string, *model.Church]
	ProjectsByUserLoader *dataloader.Loader[string, []*model.Project]
	RolesByUserLoader    *dataloader.Loader[string, []*model.UserRole]
	ProjectByIDLoader    *dataloader.Loader[string, *model.Project]
	EventByIDLoader      *dataloader.Loader[string, *model.Event]
	TeamByIDLoader       *dataloader.Loader[string, *model.Team]
}

// NewLoaders creates all dataloaders with batch functions and default caching
// Should be called once at server startup
func NewLoaders(db *database.DB, cache *cache.CacheWithRegistry) *Loaders {
	return &Loaders{
		UserByIDLoader: dataloader.NewBatchedLoader(
			userByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.User](100),
		),
		ChurchLoader: dataloader.NewBatchedLoader(
			churchBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.Church](100),
		),
		ProjectsByUserLoader: dataloader.NewBatchedLoader(
			projectsByUserBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Project](100),
		),
		RolesByUserLoader: dataloader.NewBatchedLoader(
			rolesByUserBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.UserRole](100),
		),
		ProjectByIDLoader: dataloader.NewBatchedLoader(
			projectByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.Project](100),
		),
		EventByIDLoader: dataloader.NewBatchedLoader(
			eventByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.Event](100),
		),
		TeamByIDLoader: dataloader.NewBatchedLoader(
			teamByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.Team](100),
		),
	}
}
