package loaders

import (
	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// Loaders holds all dataloader instances for batching database queries
// These are shared globally across all requests and use built-in caching
type Loaders struct {
	UserByIDLoader              *dataloader.Loader[string, *model.User]
	ChurchLoader                *dataloader.Loader[string, *model.Church]
	ProjectsByUserLoader        *dataloader.Loader[string, []*model.Project]
	EventsByUserLoader          *dataloader.Loader[string, []*model.Event]
	EventsByProjectLoader       *dataloader.Loader[string, []*model.Event]
	TeamsByUserLoader           *dataloader.Loader[string, []*model.Team]
	TeamsByProjectLoader        *dataloader.Loader[string, []*model.Team]
	TeamsBySuperTeamLoader      *dataloader.Loader[string, []*model.Team]
	SuperTeamsByUserLoader      *dataloader.Loader[string, []*model.SuperTeam]
	RolesByUserLoader           *dataloader.Loader[string, []*model.UserRole]
	UsersByTeamLoader           *dataloader.Loader[string, []*model.User]
	ProjectByIDLoader           *dataloader.Loader[string, *model.Project]
	EventByIDLoader             *dataloader.Loader[string, *model.Event]
	TeamByIDLoader              *dataloader.Loader[string, *model.Team]
	SuperTeamByIDLoader         *dataloader.Loader[string, *model.SuperTeam]
	AchievementByIDLoader       *dataloader.Loader[string, model.Achievement]
	AchievementsByProjectLoader *dataloader.Loader[string, []model.Achievement]
	ArticlesByAchievementLoader *dataloader.Loader[string, []model.Article]
	TracksByAchievementLoader   *dataloader.Loader[string, []model.Track]
	ChallengeByIDLoader         *dataloader.Loader[string, *model.Challenge]
	ChallengesByProjectLoader   *dataloader.Loader[string, []*model.Challenge]
	ChallengesByEventLoader     *dataloader.Loader[string, []*model.Challenge]
	StreakByIDLoader            *dataloader.Loader[string, *model.Streak]
	StreaksByProjectLoader      *dataloader.Loader[string, []*model.Streak]
	RelevantDaysByStreakLoader  *dataloader.Loader[string, []model.DateRange]
	UserStreakActivityLoader    *dataloader.Loader[UserStreakActivityKey, []*sqlc.UserStreakActivity]
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
		EventsByUserLoader: dataloader.NewBatchedLoader(
			eventsByUserBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Event](100),
		),
		EventsByProjectLoader: dataloader.NewBatchedLoader(
			eventsByProjectBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Event](100),
		),
		TeamsByUserLoader: dataloader.NewBatchedLoader(
			teamsByUserBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Team](100),
		),
		TeamsByProjectLoader: dataloader.NewBatchedLoader(
			teamsByProjectBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Team](100),
		),
		TeamsBySuperTeamLoader: dataloader.NewBatchedLoader(
			teamsBySuperTeamBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Team](100),
		),
		SuperTeamsByUserLoader: dataloader.NewBatchedLoader(
			superTeamsByUserBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.SuperTeam](100),
		),
		RolesByUserLoader: dataloader.NewBatchedLoader(
			rolesByUserBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.UserRole](100),
		),
		UsersByTeamLoader: dataloader.NewBatchedLoader(
			usersByTeamBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.User](100),
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
		SuperTeamByIDLoader: dataloader.NewBatchedLoader(
			superTeamByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.SuperTeam](100),
		),
		AchievementByIDLoader: dataloader.NewBatchedLoader(
			achievementByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, model.Achievement](100),
		),
		AchievementsByProjectLoader: dataloader.NewBatchedLoader(
			achievementsByProjectBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []model.Achievement](100),
		),
		ArticlesByAchievementLoader: dataloader.NewBatchedLoader(
			articlesByAchievementBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []model.Article](100),
		),
		TracksByAchievementLoader: dataloader.NewBatchedLoader(
			tracksByAchievementBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []model.Track](100),
		),
		ChallengeByIDLoader: dataloader.NewBatchedLoader(
			challengeByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.Challenge](100),
		),
		ChallengesByProjectLoader: dataloader.NewBatchedLoader(
			challengesByProjectBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Challenge](100),
		),
		ChallengesByEventLoader: dataloader.NewBatchedLoader(
			challengesByEventBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Challenge](100),
		),
		StreakByIDLoader: dataloader.NewBatchedLoader(
			streakByIDBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, *model.Streak](100),
		),
		StreaksByProjectLoader: dataloader.NewBatchedLoader(
			streaksByProjectBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []*model.Streak](100),
		),
		RelevantDaysByStreakLoader: dataloader.NewBatchedLoader(
			relevantDaysByStreakBatchFunc(db, cache),
			dataloader.WithBatchCapacity[string, []model.DateRange](100),
		),
		UserStreakActivityLoader: dataloader.NewBatchedLoader(
			userStreakActivityBatchFunc(db, cache),
			dataloader.WithBatchCapacity[UserStreakActivityKey, []*sqlc.UserStreakActivity](100),
		),
	}
}
