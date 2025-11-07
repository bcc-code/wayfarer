package loaders

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/graph-gophers/dataloader/v7"
)

// achievementByIDBatchFunc batches loading achievements by IDs
func achievementByIDBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []string) []*dataloader.Result[model.Achievement] {
	return func(ctx context.Context, ids []string) []*dataloader.Result[model.Achievement] {
		// Check cache first for each ID
		achievementMap := make(map[string]model.Achievement)
		missingIDs := []string{}

		for _, id := range ids {
			cacheKey := cache.AchievementKey(id)
			if cached, ok := c.Get(cacheKey); ok {
				if achievement, ok := cached.(model.Achievement); ok {
					achievementMap[id] = achievement
					continue
				}
			}
			missingIDs = append(missingIDs, id)
		}

		// Query database only for cache misses
		if len(missingIDs) > 0 {
			rows, err := db.Queries.GetAchievementsByIDs(ctx, missingIDs)
			if err != nil {
				// Return error for all IDs
				results := make([]*dataloader.Result[model.Achievement], len(ids))
				for i := range results {
					results[i] = &dataloader.Result[model.Achievement]{Error: err}
				}
				return results
			}

			// Convert to GraphQL model and populate cache
			for _, row := range rows {
				var achievement model.Achievement
				var err error

				hidden := false
				if row.Hidden != nil {
					hidden = *row.Hidden
				}

				switch row.AchievementType {
				case "SIMPLE":
					achievement, err = convertToSimpleAchievement(row, hidden)
				case "READING":
					achievement, err = convertToReadingAchievement(row, hidden)
				case "LISTENING":
					achievement, err = convertToListeningAchievement(row, hidden)
				case "STREAK":
					achievement, err = convertToStreakAchievement(row, hidden)
				default:
					err = fmt.Errorf("unknown achievement type: %s", row.AchievementType)
				}

				if err != nil {
					// If we can't convert one achievement, return error for all
					results := make([]*dataloader.Result[model.Achievement], len(ids))
					for i := range results {
						results[i] = &dataloader.Result[model.Achievement]{Error: err}
					}
					return results
				}

				achievementMap[row.ID] = achievement
				// Store in cache with default TTL (15 minutes)
				c.Set(cache.AchievementKey(row.ID), achievement)
			}
		}

		// Return results in same order as input IDs
		results := make([]*dataloader.Result[model.Achievement], len(ids))
		for i, id := range ids {
			if achievement, ok := achievementMap[id]; ok {
				results[i] = &dataloader.Result[model.Achievement]{Data: achievement}
			} else {
				results[i] = &dataloader.Result[model.Achievement]{
					Error: fmt.Errorf("achievement not found: %s", id),
				}
			}
		}
		return results
	}
}

// Helper functions to convert database rows to GraphQL models

func convertToSimpleAchievement(row *sqlc.GetAchievementsByIDsRow, hidden bool) (model.Achievement, error) {
	return &model.SimpleAchievement{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Image:       row.ImageUrl,
		Points:      int(row.Points),
		Hidden:      hidden,
		ProjectID:   row.ProjectID,
		EventID:     row.EventID,
		ChallengeID: row.ChallengeID,
	}, nil
}

func convertToReadingAchievement(row *sqlc.GetAchievementsByIDsRow, hidden bool) (model.Achievement, error) {
	// Parse the reading articles JSON
	var articlesData []map[string]interface{}
	if row.ReadingArticles != nil {
		jsonBytes, err := json.Marshal(row.ReadingArticles)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal reading articles: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &articlesData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal reading articles: %w", err)
		}
	}

	articles := make([]model.Article, 0, len(articlesData))
	for _, articleData := range articlesData {
		article := model.Article{
			ID:     articleData["id"].(string),
			Title:  articleData["title"].(string),
			Author: articleData["author"].(string),
			URL:    articleData["url"].(string),
		}
		articles = append(articles, article)
	}

	return &model.ReadingAchievement{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Image:       row.ImageUrl,
		Points:      int(row.Points),
		Hidden:      hidden,
		ProjectID:   row.ProjectID,
		EventID:     row.EventID,
		ChallengeID: row.ChallengeID,
		Articles:    articles,
		UserHasRead: []model.Article{}, // Will be populated by resolver if needed
		NextArticle: nil,               // Will be populated by resolver if needed
	}, nil
}

func convertToListeningAchievement(row *sqlc.GetAchievementsByIDsRow, hidden bool) (model.Achievement, error) {
	// Parse the listening tracks JSON
	var tracksData []map[string]interface{}
	if row.ListeningTracks != nil {
		jsonBytes, err := json.Marshal(row.ListeningTracks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal listening tracks: %w", err)
		}
		if err := json.Unmarshal(jsonBytes, &tracksData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal listening tracks: %w", err)
		}
	}

	tracks := make([]model.Track, 0, len(tracksData))
	for _, trackData := range tracksData {
		description := ""
		if desc, ok := trackData["description"]; ok && desc != nil {
			description = desc.(string)
		}
		imageURL := ""
		if img, ok := trackData["image_url"]; ok && img != nil {
			imageURL = img.(string)
		}

		track := model.Track{
			ID:          trackData["id"].(string),
			Name:        trackData["name"].(string),
			Description: description,
			Image:       imageURL,
		}
		tracks = append(tracks, track)
	}

	return &model.ListeningAchievement{
		ID:              row.ID,
		Name:            row.Name,
		Description:     row.Description,
		Image:           row.ImageUrl,
		Points:          int(row.Points),
		Hidden:          hidden,
		ProjectID:       row.ProjectID,
		EventID:         row.EventID,
		ChallengeID:     row.ChallengeID,
		Tracks:          tracks,
		UserHasListened: []model.Track{}, // Will be populated by resolver if needed
		NextTrack:       nil,             // Will be populated by resolver if needed
	}, nil
}

func convertToStreakAchievement(row *sqlc.GetAchievementsByIDsRow, hidden bool) (model.Achievement, error) {
	if row.StreakID == nil || row.NeededStreak == nil {
		return nil, fmt.Errorf("streak achievement missing required fields: streak_id or needed_streak")
	}

	return &model.StreakAchievement{
		ID:           row.ID,
		Name:         row.Name,
		Description:  row.Description,
		Image:        row.ImageUrl,
		Points:       int(row.Points),
		Hidden:       hidden,
		ProjectID:    row.ProjectID,
		EventID:      row.EventID,
		ChallengeID:  row.ChallengeID,
		StreakID:     *row.StreakID,
		NeededStreak: int(*row.NeededStreak),
		Streak:       nil, // Will be populated by resolver
	}, nil
}
