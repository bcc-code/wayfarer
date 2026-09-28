package translations

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/bcc-media/wayfarer/common"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/orsinium-labs/enum"
)

// TranslationsProvider is the interface for translation providers (e.g., Phrase)
type TranslationsProvider interface {
	SendToTranslation(ctx context.Context, collection string, data []common.TranslationData) error
	ProcessWebhook(ctx context.Context, originalRequest *http.Request, hookData []byte) (*TranslatableCollection, []common.TranslationData, error)
}

// Service handles translation exports and imports
type Service struct {
	provider TranslationsProvider
	queries  *sqlc.Queries
}

// TranslatableCollection represents a collection of translatable entities
type TranslatableCollection enum.Member[string]

var (
	CollectionProjects     = TranslatableCollection{"projects"}
	CollectionEvents       = TranslatableCollection{"events"}
	CollectionChallenges   = TranslatableCollection{"challenges"}
	CollectionAchievements = TranslatableCollection{"achievements"}
	CollectionQuizzes      = TranslatableCollection{"quizzes"}
	CollectionConsents     = TranslatableCollection{"consents"}

	TranslatableCollections = enum.New(
		CollectionProjects,
		CollectionEvents,
		CollectionChallenges,
		CollectionAchievements,
		CollectionQuizzes,
		CollectionConsents,
	)
)

// NewService creates a new translation service
func NewService(queries *sqlc.Queries, provider TranslationsProvider) *Service {
	return &Service{
		provider: provider,
		queries:  queries,
	}
}

// SendAllToTranslation exports all translatable collections to the translation provider
func (s *Service) SendAllToTranslation(ctx context.Context) []error {
	errs := make([]error, 0)
	for _, collection := range TranslatableCollections.Members() {
		if err := s.SendCollectionToTranslation(ctx, collection); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// SendCollectionToTranslation exports a specific collection to the translation provider
func (s *Service) SendCollectionToTranslation(ctx context.Context, collection TranslatableCollection) error {
	var err error
	var data []common.TranslationData

	switch collection {
	case CollectionProjects:
		data, err = s.getDataForProjects(ctx)
	case CollectionEvents:
		data, err = s.getDataForEvents(ctx)
	case CollectionChallenges:
		data, err = s.getDataForChallenges(ctx)
	case CollectionAchievements:
		data, err = s.getDataForAchievements(ctx)
	case CollectionQuizzes:
		data, err = s.getDataForQuizzes(ctx)
	case CollectionConsents:
		data, err = s.getDataForConsents(ctx)
	}

	if err != nil {
		return err
	}

	return s.sendToProviderIfNeeded(ctx, collection, data)
}

// sendToProviderIfNeeded sends data to provider only if the hash has changed
func (s *Service) sendToProviderIfNeeded(ctx context.Context, collection TranslatableCollection, data []common.TranslationData) error {
	if len(data) == 0 {
		return nil
	}

	// Sort for consistent hashing
	sort.Slice(data, func(i, j int) bool {
		return data[i].ID < data[j].ID
	})

	marshalledData, _ := json.Marshal(data)
	h := sha1.New()
	h.Write(marshalledData)
	hash := h.Sum(nil)

	// Check if this hash was already sent
	existingHash, err := s.queries.GetTranslationHash(ctx, collection.Value)
	if err == nil && string(existingHash) == string(hash) {
		// No changes, skip sending
		return nil
	}

	// Send to provider
	err = s.provider.SendToTranslation(ctx, collection.Value, data)
	if err != nil {
		return err
	}

	// Update hash
	_ = s.queries.UpsertTranslationHash(ctx, sqlc.UpsertTranslationHashParams{
		Collection: collection.Value,
		Hash:       hash,
	})

	return nil
}

// UpdateTranslations imports translations from the provider into the database
func (s *Service) UpdateTranslations(ctx context.Context, collection *TranslatableCollection, data []common.TranslationData) []error {
	if collection == nil {
		return nil
	}

	switch *collection {
	case CollectionProjects:
		return s.updateProjects(ctx, data)
	case CollectionEvents:
		return s.updateEvents(ctx, data)
	case CollectionChallenges:
		return s.updateChallenges(ctx, data)
	case CollectionAchievements:
		return s.updateAchievements(ctx, data)
	case CollectionQuizzes:
		return s.updateQuizzes(ctx, data)
	case CollectionConsents:
		return s.updateConsents(ctx, data)
	}
	return nil
}

// ProcessWebhook processes an incoming webhook from the translation provider
func (s *Service) ProcessWebhook(ctx context.Context, r *http.Request, body []byte) (*TranslatableCollection, []common.TranslationData, error) {
	return s.provider.ProcessWebhook(ctx, r, body)
}

// mustToJSON marshals a value to JSON, panicking on error (should never happen with our types)
func mustToJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
