package loaders

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/cache"
	"github.com/bcc-media/wayfarer/internal/database"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/graph-gophers/dataloader/v7"
)

// TranslationKey is the composite key for translation lookups
type TranslationKey struct {
	EntityType string // "project", "event", "team", etc.
	EntityID   string
	LangCode   string
}

func (k TranslationKey) String() string {
	return fmt.Sprintf("%s:%s:%s", k.EntityType, k.EntityID, k.LangCode)
}

// Translation holds translated fields for an entity
// Fields are pointers - nil means "use base value"
type Translation struct {
	EntityID             string
	LangCode             string
	Name                 *string
	Description          *string // Used by most entity types
	DescriptionPending   *string // Only for achievements
	DescriptionCompleted *string // Only for achievements
	NotificationText     *string // Only for achievements
	Rules                *string // Only for projects
	ButtonText           *string // Only for challenges
	Title                *string // Only for articles and consents
	ShortText            *string // Only for consents
	Author               *string // Only for articles
	Body                 *string // Only for consents
	QuestionText         *string // Only for quiz questions
	AnswerText           *string // Only for quiz answers
}

// translationBatchFunc batches loading translations by entity type, ID, and language code
func translationBatchFunc(db *database.DB, c *cache.CacheWithRegistry) func(context.Context, []TranslationKey) []*dataloader.Result[*Translation] {
	return func(ctx context.Context, keys []TranslationKey) []*dataloader.Result[*Translation] {
		// Group keys by entity type and language
		translationMap := make(map[string]*Translation)
		missingByType := make(map[string][]TranslationKey) // entityType -> keys

		// Check cache first
		for _, key := range keys {
			cacheKey := cache.TranslationKey(key.EntityType, key.EntityID, key.LangCode)
			if cached, ok := c.Get(cacheKey); ok {
				if trans, ok := cached.(*Translation); ok {
					translationMap[key.String()] = trans
					continue
				}
			}
			missingByType[key.EntityType] = append(missingByType[key.EntityType], key)
		}

		// Query database for cache misses, grouped by entity type
		for entityType, typeKeys := range missingByType {
			// Group by language code for efficient queries
			byLang := make(map[string][]string) // langCode -> entityIDs
			for _, key := range typeKeys {
				byLang[key.LangCode] = append(byLang[key.LangCode], key.EntityID)
			}

			for langCode, entityIDs := range byLang {
				translations, err := queryTranslations(ctx, db, entityType, entityIDs, langCode)
				if err != nil {
					// Log error but continue - translations are optional
					continue
				}

				// Map results
				for _, trans := range translations {
					key := TranslationKey{EntityType: entityType, EntityID: trans.EntityID, LangCode: langCode}
					translationMap[key.String()] = trans
					c.Set(cache.TranslationKey(entityType, trans.EntityID, langCode), trans)
				}

				// Store empty translations for entities without translations (negative cache)
				foundIDs := make(map[string]bool)
				for _, trans := range translations {
					foundIDs[trans.EntityID] = true
				}
				for _, entityID := range entityIDs {
					if !foundIDs[entityID] {
						key := TranslationKey{EntityType: entityType, EntityID: entityID, LangCode: langCode}
						emptyTrans := &Translation{EntityID: entityID, LangCode: langCode}
						translationMap[key.String()] = emptyTrans
						c.Set(cache.TranslationKey(entityType, entityID, langCode), emptyTrans)
					}
				}
			}
		}

		// Return results in same order as input
		results := make([]*dataloader.Result[*Translation], len(keys))
		for i, key := range keys {
			if trans, ok := translationMap[key.String()]; ok {
				results[i] = &dataloader.Result[*Translation]{Data: trans}
			} else {
				// No translation found - return empty translation (not an error)
				results[i] = &dataloader.Result[*Translation]{Data: &Translation{EntityID: key.EntityID, LangCode: key.LangCode}}
			}
		}
		return results
	}
}

// queryTranslations queries the appropriate translation table based on entity type
func queryTranslations(ctx context.Context, db *database.DB, entityType string, entityIDs []string, langCode string) ([]*Translation, error) {
	switch entityType {
	case "project":
		rows, err := db.Queries.GetProjectTranslationsByIDs(ctx, sqlc.GetProjectTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:    row.ProjectID,
				LangCode:    row.LanguageCode,
				Name:        row.Name,
				Description: row.Description,
				Rules:       row.Rules,
			}
		}
		return translations, nil

	case "event":
		rows, err := db.Queries.GetEventTranslationsByIDs(ctx, sqlc.GetEventTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:    row.EventID,
				LangCode:    row.LanguageCode,
				Name:        row.Name,
				Description: row.Description,
			}
		}
		return translations, nil

	case "team":
		rows, err := db.Queries.GetTeamTranslationsByIDs(ctx, sqlc.GetTeamTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:    row.TeamID,
				LangCode:    row.LanguageCode,
				Name:        row.Name,
				Description: row.Description,
			}
		}
		return translations, nil

	case "superteam":
		rows, err := db.Queries.GetSuperTeamTranslationsByIDs(ctx, sqlc.GetSuperTeamTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:    row.SuperTeamID,
				LangCode:    row.LanguageCode,
				Name:        row.Name,
				Description: row.Description,
			}
		}
		return translations, nil

	case "streak":
		rows, err := db.Queries.GetStreakTranslationsByIDs(ctx, sqlc.GetStreakTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:    row.StreakID,
				LangCode:    row.LanguageCode,
				Name:        row.Name,
				Description: row.Description,
			}
		}
		return translations, nil

	case "challenge":
		rows, err := db.Queries.GetChallengeTranslationsByIDs(ctx, sqlc.GetChallengeTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:    row.ChallengeID,
				LangCode:    row.LanguageCode,
				Name:        row.Name,
				Description: row.Description,
				ButtonText:  row.ButtonText,
			}
		}
		return translations, nil

	case "achievement":
		rows, err := db.Queries.GetAchievementTranslationsByIDs(ctx, sqlc.GetAchievementTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:             row.AchievementID,
				LangCode:             row.LanguageCode,
				Name:                 row.Name,
				DescriptionPending:   row.DescriptionPending,
				DescriptionCompleted: row.DescriptionCompleted,
				NotificationText:     row.NotificationText,
			}
		}
		return translations, nil

	case "consent":
		rows, err := db.Queries.GetConsentTranslationsByIDs(ctx, sqlc.GetConsentTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:   row.ConsentID,
				LangCode:   row.LanguageCode,
				Title:      row.Title,
				ShortText:  row.ShortText,
				Body:       row.Body,
				ButtonText: row.ButtonText,
			}
		}
		return translations, nil

	case "quiz":
		rows, err := db.Queries.GetQuizTranslationsByIDs(ctx, sqlc.GetQuizTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:    row.QuizID,
				LangCode:    row.LanguageCode,
				Name:        row.Name,
				Description: row.Description,
			}
		}
		return translations, nil

	case "quiz_question":
		rows, err := db.Queries.GetQuizQuestionTranslationsByIDs(ctx, sqlc.GetQuizQuestionTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:     row.QuestionID,
				LangCode:     row.LanguageCode,
				QuestionText: row.QuestionText,
			}
		}
		return translations, nil

	case "quiz_answer":
		rows, err := db.Queries.GetQuizAnswerTranslationsByIDs(ctx, sqlc.GetQuizAnswerTranslationsByIDsParams{
			EntityIds:    entityIDs,
			LanguageCode: langCode,
		})
		if err != nil {
			return nil, err
		}
		translations := make([]*Translation, len(rows))
		for i, row := range rows {
			translations[i] = &Translation{
				EntityID:   row.AnswerID,
				LangCode:   row.LanguageCode,
				AnswerText: row.AnswerText,
			}
		}
		return translations, nil

	default:
		return nil, fmt.Errorf("unknown entity type for translation: %s", entityType)
	}
}
