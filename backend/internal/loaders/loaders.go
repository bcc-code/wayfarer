package loaders

import (
	"context"
	"time"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// Loaders holds all dataloader instances for batching database queries
// These are shared globally across all requests and rely on Ristretto cache for data caching
type Loaders struct {
	UserByIDLoader                 *dataloader.Loader[string, *model.User]
	ChurchLoader                   *dataloader.Loader[string, *model.Church]
	ProjectsByUserLoader           *dataloader.Loader[string, []*model.Project]
	EventsByUserLoader             *dataloader.Loader[string, []*model.Event]
	EventsByProjectLoader          *dataloader.Loader[string, []*model.Event]
	TeamsByUserLoader              *dataloader.Loader[string, []*model.Team]
	TeamsByProjectLoader           *dataloader.Loader[string, []*model.Team]
	TeamsBySuperTeamLoader         *dataloader.Loader[string, []*model.Team]
	SuperTeamsByUserLoader         *dataloader.Loader[string, []*model.SuperTeam]
	RolesByUserLoader              *dataloader.Loader[string, []*model.UserRole]
	UsersByTeamLoader              *dataloader.Loader[string, []*model.TeamMember]
	TeamMemberLeaderboardLoader    *dataloader.Loader[string, []model.LeaderboardEntry]
	ProjectByIDLoader              *dataloader.Loader[string, *model.Project]
	EventByIDLoader                *dataloader.Loader[string, *model.Event]
	TeamByIDLoader                 *dataloader.Loader[string, *model.Team]
	SuperTeamByIDLoader            *dataloader.Loader[string, *model.SuperTeam]
	AchievementByIDLoader          *dataloader.Loader[string, model.Achievement]
	AchievementsByProjectLoader    *dataloader.Loader[string, []model.Achievement]
	ArticlesByAchievementLoader    *dataloader.Loader[string, []model.Article]
	TracksByAchievementLoader      *dataloader.Loader[string, []model.Track]
	ChallengeByIDLoader            *dataloader.Loader[string, *model.Challenge]
	ChallengesByProjectLoader      *dataloader.Loader[string, []*model.Challenge]
	ChallengesByEventLoader        *dataloader.Loader[string, []*model.Challenge]
	StreakByIDLoader               *dataloader.Loader[string, *model.Streak]
	StreaksByProjectLoader         *dataloader.Loader[string, []*model.Streak]
	RelevantDaysByStreakLoader     *dataloader.Loader[string, []model.DateRange]
	UserStreakActivityLoader         *dataloader.Loader[UserStreakActivityKey, []*sqlc.UserStreakActivity]
	UserAchievementTimestampLoader   *dataloader.Loader[UserAchievementKey, *time.Time]
	TranslationLoader                *dataloader.Loader[TranslationKey, *Translation]
	ConsentByIDLoader                *dataloader.Loader[string, *model.Consent]
}

// newBatchedLoader creates a new batched dataloader with standard configuration:
// - Batch capacity of 100
// - Cache disabled (we rely on Ristretto cache in batch functions instead)
func newBatchedLoader[K comparable, V any](
	batchFunc func(context.Context, []K) []*dataloader.Result[V],
) *dataloader.Loader[K, V] {
	return dataloader.NewBatchedLoader(
		batchFunc,
		dataloader.WithBatchCapacity[K, V](100),
		dataloader.WithCache[K, V](&dataloader.NoCache[K, V]{}), // Disable internal cache, use Ristretto instead
	)
}

// NewLoaders creates all dataloaders with batch functions
// Dataloaders provide request batching while relying on Ristretto cache for data caching
// Should be called once at server startup
func NewLoaders(db *database.DB, cache *cache.CacheWithRegistry) *Loaders {
	return &Loaders{
		UserByIDLoader:                 newBatchedLoader(userByIDBatchFunc(db, cache)),
		ChurchLoader:                   newBatchedLoader(churchBatchFunc(db, cache)),
		ProjectsByUserLoader:           newBatchedLoader(projectsByUserBatchFunc(db, cache)),
		EventsByUserLoader:             newBatchedLoader(eventsByUserBatchFunc(db, cache)),
		EventsByProjectLoader:          newBatchedLoader(eventsByProjectBatchFunc(db, cache)),
		TeamsByUserLoader:              newBatchedLoader(teamsByUserBatchFunc(db, cache)),
		TeamsByProjectLoader:           newBatchedLoader(teamsByProjectBatchFunc(db, cache)),
		TeamsBySuperTeamLoader:         newBatchedLoader(teamsBySuperTeamBatchFunc(db, cache)),
		SuperTeamsByUserLoader:         newBatchedLoader(superTeamsByUserBatchFunc(db, cache)),
		RolesByUserLoader:              newBatchedLoader(rolesByUserBatchFunc(db, cache)),
		UsersByTeamLoader:              newBatchedLoader(usersByTeamBatchFunc(db, cache)),
		TeamMemberLeaderboardLoader:    newBatchedLoader(teamMemberLeaderboardBatchFunc(db, cache)),
		ProjectByIDLoader:              newBatchedLoader(projectByIDBatchFunc(db, cache)),
		EventByIDLoader:                newBatchedLoader(eventByIDBatchFunc(db, cache)),
		TeamByIDLoader:                 newBatchedLoader(teamByIDBatchFunc(db, cache)),
		SuperTeamByIDLoader:            newBatchedLoader(superTeamByIDBatchFunc(db, cache)),
		AchievementByIDLoader:          newBatchedLoader(achievementByIDBatchFunc(db, cache)),
		AchievementsByProjectLoader:    newBatchedLoader(achievementsByProjectBatchFunc(db, cache)),
		ArticlesByAchievementLoader:    newBatchedLoader(articlesByAchievementBatchFunc(db, cache)),
		TracksByAchievementLoader:      newBatchedLoader(tracksByAchievementBatchFunc(db, cache)),
		ChallengeByIDLoader:            newBatchedLoader(challengeByIDBatchFunc(db, cache)),
		ChallengesByProjectLoader:      newBatchedLoader(challengesByProjectBatchFunc(db, cache)),
		ChallengesByEventLoader:        newBatchedLoader(challengesByEventBatchFunc(db, cache)),
		StreakByIDLoader:               newBatchedLoader(streakByIDBatchFunc(db, cache)),
		StreaksByProjectLoader:         newBatchedLoader(streaksByProjectBatchFunc(db, cache)),
		RelevantDaysByStreakLoader:     newBatchedLoader(relevantDaysByStreakBatchFunc(db, cache)),
		UserStreakActivityLoader:         newBatchedLoader(userStreakActivityBatchFunc(db, cache)),
		UserAchievementTimestampLoader:   newBatchedLoader(userAchievementTimestampBatchFunc(db, cache)),
		TranslationLoader:                newBatchedLoader(translationBatchFunc(db, cache)),
		ConsentByIDLoader:                newBatchedLoader(consentByIDBatchFunc(db, cache)),
	}
}
