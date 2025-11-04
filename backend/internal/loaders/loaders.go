package loaders

import (
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
}

// NewLoaders creates all dataloaders with batch functions and default caching
// Should be called once at server startup
func NewLoaders(db *database.DB) *Loaders {
	return &Loaders{
		UserByIDLoader: dataloader.NewBatchedLoader(
			userByIDBatchFunc(db),
			dataloader.WithBatchCapacity[string, *model.User](100),
		),
		ChurchLoader: dataloader.NewBatchedLoader(
			churchBatchFunc(db),
			dataloader.WithBatchCapacity[string, *model.Church](100),
		),
		ProjectsByUserLoader: dataloader.NewBatchedLoader(
			projectsByUserBatchFunc(db),
			dataloader.WithBatchCapacity[string, []*model.Project](100),
		),
	}
}
