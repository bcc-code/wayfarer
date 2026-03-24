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
	UserByIDLoader                           *dataloader.Loader[string, *model.User]
	ChurchLoader                             *dataloader.Loader[string, *model.Church]
	ProjectsByUserLoader                     *dataloader.Loader[string, []*model.Project]
	EventsByUserLoader                       *dataloader.Loader[string, []*model.Event]
	EventsByProjectLoader                    *dataloader.Loader[string, []*model.Event]
	TeamsByUserLoader                        *dataloader.Loader[string, []*model.Team]
	TeamsByProjectLoader                     *dataloader.Loader[string, []*model.Team]
	TeamsBySuperTeamLoader                   *dataloader.Loader[string, []*model.Team]
	SuperTeamsByUserLoader                   *dataloader.Loader[string, []*model.SuperTeam]
	RolesByUserLoader                        *dataloader.Loader[string, []*model.UserRole]
	UsersByTeamLoader                        *dataloader.Loader[string, []*model.TeamMember]
	TeamMemberLeaderboardLoader              *dataloader.Loader[string, []model.LeaderboardEntry]
	ProjectByIDLoader                        *dataloader.Loader[string, *model.Project]
	EventByIDLoader                          *dataloader.Loader[string, *model.Event]
	TeamByIDLoader                           *dataloader.Loader[string, *model.Team]
	SuperTeamByIDLoader                      *dataloader.Loader[string, *model.SuperTeam]
	AchievementByIDLoader                    *dataloader.Loader[string, model.Achievement]
	AchievementsByProjectLoader              *dataloader.Loader[string, []model.Achievement]
	ContentItemsByAchievementLoader          *dataloader.Loader[string, []*model.ContentItem]
	ChallengeByIDLoader                      *dataloader.Loader[string, model.Challenge]
	ChallengesByProjectLoader                *dataloader.Loader[string, []model.Challenge]
	ChallengesByEventLoader                  *dataloader.Loader[string, []model.Challenge]
	StreakItemsByAchievementLoader           *dataloader.Loader[string, []*model.ContentItem]
	UserStreakProgressLoader                 *dataloader.Loader[UserAchievementKey, []*sqlc.UserStreakProgress]
	UserContentProgressLoader                *dataloader.Loader[UserAchievementKey, []*sqlc.UserContentProgress]
	UserAchievementTimestampLoader           *dataloader.Loader[UserAchievementKey, *time.Time]
	UserAchievementCelebratedTimestampLoader *dataloader.Loader[UserAchievementKey, *time.Time]
	UserChallengeCompletionTimestampLoader   *dataloader.Loader[UserChallengeKey, *time.Time]
	UserChallengeEnrollmentTimestampLoader   *dataloader.Loader[UserChallengeKey, *time.Time]
	TranslationLoader                        *dataloader.Loader[TranslationKey, *Translation]
	ConsentByIDLoader                        *dataloader.Loader[string, *model.Consent]
	QuizByIDLoader                           *dataloader.Loader[string, *model.Quiz]
	QuizByChallengeIDLoader                  *dataloader.Loader[string, *model.Quiz]
	QuizQuestionsByQuizLoader                *dataloader.Loader[string, []model.QuizQuestion]
	QuizAnswersByQuestionLoader              *dataloader.Loader[string, []*model.QuizPredefinedAnswer]
	QuizSubmissionsByUserLoader              *dataloader.Loader[string, []*model.QuizSubmission]
	QuizResponsesBySubmissionLoader          *dataloader.Loader[string, []model.QuizResponse]
	QuizSessionByIDLoader                    *dataloader.Loader[string, *sqlc.QuizSession]
	UserIDsByTeamLoader                      *dataloader.Loader[string, []string]
	UserIDsBySuperTeamLoader                 *dataloader.Loader[string, []string]
	UserIDsByChurchInProjectLoader           *dataloader.Loader[ChurchProjectKey, []string]
	UserIDsInProjectLoader                   *dataloader.Loader[string, []string]
	ExternalContentByIDLoader                *dataloader.Loader[string, *model.ExternalContent]
	ExternalContentTranslationsLoader        *dataloader.Loader[string, []model.ExternalContentTranslation]
	ImageMetadataByURLLoader                 *dataloader.Loader[string, *model.Image]
	ScoreJournalByIDLoader                   *dataloader.Loader[string, *model.ScoreJournal]
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
		UserByIDLoader:                           newBatchedLoader(userByIDBatchFunc(db, cache)),
		ChurchLoader:                             newBatchedLoader(churchBatchFunc(db, cache)),
		ProjectsByUserLoader:                     newBatchedLoader(projectsByUserBatchFunc(db, cache)),
		EventsByUserLoader:                       newBatchedLoader(eventsByUserBatchFunc(db, cache)),
		EventsByProjectLoader:                    newBatchedLoader(eventsByProjectBatchFunc(db, cache)),
		TeamsByUserLoader:                        newBatchedLoader(teamsByUserBatchFunc(db, cache)),
		TeamsByProjectLoader:                     newBatchedLoader(teamsByProjectBatchFunc(db, cache)),
		TeamsBySuperTeamLoader:                   newBatchedLoader(teamsBySuperTeamBatchFunc(db, cache)),
		SuperTeamsByUserLoader:                   newBatchedLoader(superTeamsByUserBatchFunc(db, cache)),
		RolesByUserLoader:                        newBatchedLoader(rolesByUserBatchFunc(db, cache)),
		UsersByTeamLoader:                        newBatchedLoader(usersByTeamBatchFunc(db, cache)),
		TeamMemberLeaderboardLoader:              newBatchedLoader(teamMemberLeaderboardBatchFunc(db, cache)),
		ProjectByIDLoader:                        newBatchedLoader(projectByIDBatchFunc(db, cache)),
		EventByIDLoader:                          newBatchedLoader(eventByIDBatchFunc(db, cache)),
		TeamByIDLoader:                           newBatchedLoader(teamByIDBatchFunc(db, cache)),
		SuperTeamByIDLoader:                      newBatchedLoader(superTeamByIDBatchFunc(db, cache)),
		AchievementByIDLoader:                    newBatchedLoader(achievementByIDBatchFunc(db, cache)),
		AchievementsByProjectLoader:              newBatchedLoader(achievementsByProjectBatchFunc(db, cache)),
		ContentItemsByAchievementLoader:          newBatchedLoader(contentItemsByAchievementBatchFunc(db, cache)),
		ChallengeByIDLoader:                      newBatchedLoader(challengeByIDBatchFunc(db, cache)),
		ChallengesByProjectLoader:                newBatchedLoader(challengesByProjectBatchFunc(db, cache)),
		ChallengesByEventLoader:                  newBatchedLoader(challengesByEventBatchFunc(db, cache)),
		StreakItemsByAchievementLoader:           newBatchedLoader(streakItemsByAchievementBatchFunc(db, cache)),
		UserStreakProgressLoader:                 newBatchedLoader(userStreakProgressBatchFunc(db, cache)),
		UserContentProgressLoader:                newBatchedLoader(userContentProgressBatchFunc(db, cache)),
		UserAchievementTimestampLoader:           newBatchedLoader(userAchievementTimestampBatchFunc(db, cache)),
		UserAchievementCelebratedTimestampLoader: newBatchedLoader(userAchievementCelebratedTimestampBatchFunc(db, cache)),
		UserChallengeCompletionTimestampLoader:   newBatchedLoader(userChallengeCompletionTimestampBatchFunc(db, cache)),
		UserChallengeEnrollmentTimestampLoader:   newBatchedLoader(userChallengeEnrollmentTimestampBatchFunc(db, cache)),
		TranslationLoader:                        newBatchedLoader(translationBatchFunc(db, cache)),
		ConsentByIDLoader:                        newBatchedLoader(consentByIDBatchFunc(db, cache)),
		QuizByIDLoader:                           newBatchedLoader(quizByIDBatchFunc(db, cache)),
		QuizByChallengeIDLoader:                  newBatchedLoader(quizByChallengeIDBatchFunc(db, cache)),
		QuizQuestionsByQuizLoader:                newBatchedLoader(quizQuestionsByQuizBatchFunc(db, cache)),
		QuizAnswersByQuestionLoader:              newBatchedLoader(quizAnswersByQuestionBatchFunc(db, cache)),
		QuizSubmissionsByUserLoader:              newBatchedLoader(quizSubmissionsByUserBatchFunc(db, cache)),
		QuizResponsesBySubmissionLoader:          newBatchedLoader(quizResponsesBySubmissionBatchFunc(db, cache)),
		QuizSessionByIDLoader:                    newBatchedLoader(quizSessionByIDBatchFunc(db, cache)),
		UserIDsByTeamLoader:                      newBatchedLoader(userIDsByTeamBatchFunc(db, cache)),
		UserIDsBySuperTeamLoader:                 newBatchedLoader(userIDsBySuperTeamBatchFunc(db, cache)),
		UserIDsByChurchInProjectLoader:           newBatchedLoader(userIDsByChurchInProjectBatchFunc(db, cache)),
		UserIDsInProjectLoader:                   newBatchedLoader(userIDsInProjectBatchFunc(db, cache)),
		ExternalContentByIDLoader:                newBatchedLoader(externalContentByIDBatchFunc(db, cache)),
		ExternalContentTranslationsLoader:        newBatchedLoader(externalContentTranslationsBatchFunc(db, cache)),
		ImageMetadataByURLLoader:                 newBatchedLoader(imageMetadataByURLBatchFunc(db, cache)),
		ScoreJournalByIDLoader:                   newBatchedLoader(scoreJournalByIDBatchFunc(db, cache)),
	}
}
