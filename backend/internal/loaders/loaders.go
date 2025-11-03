package loaders

import (
	"context"

	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/graph/user/model"
	"github.com/graph-gophers/dataloader/v7"
)

type contextKey string

const loadersKey contextKey = "dataloaders"

// Loaders holds all dataloader instances for batching database queries
type Loaders struct {
	ChurchLoader         *dataloader.Loader[string, *model.Church]
	ProjectsByUserLoader *dataloader.Loader[string, []*model.Project]
}

// NewLoaders creates all dataloaders with batch functions
func NewLoaders(db *database.DB) *Loaders {
	return &Loaders{
		ChurchLoader: dataloader.NewBatchedLoader(
			churchBatchFunc(db),
			dataloader.WithCache[string, *model.Church](&dataloader.NoCache[string, *model.Church]{}),
		),
		ProjectsByUserLoader: dataloader.NewBatchedLoader(
			projectsByUserBatchFunc(db),
			dataloader.WithCache[string, []*model.Project](&dataloader.NoCache[string, []*model.Project]{}),
		),
	}
}

// GetLoaders retrieves loaders from context
func GetLoaders(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}

// Inject adds loaders to context
func Inject(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, loadersKey, loaders)
}
