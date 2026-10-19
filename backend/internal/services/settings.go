package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrSettingNotFound    = errors.New("setting not found")
	ErrInvalidSettingType = errors.New("invalid setting type")
	ErrInvalidSettingVal  = errors.New("invalid setting value")
	ErrProjectNotFound    = errors.New("project not found")
	ErrSettingNotEditable = errors.New("setting is not editable")
)

const (
	// SettingCurrentProjectID holds the project every end user sees.
	SettingCurrentProjectID = "current_project_id"
	// SettingFrontendConfig holds the JSON blob Query.frontendConfig serves.
	SettingFrontendConfig = "frontend_config"
)

// appDataSettings are keys the application reads directly rather than through
// config. They override no environment variable and need no restart, but they
// are owned by the app and editable.
var appDataSettings = map[string]bool{
	SettingCurrentProjectID: true,
	SettingFrontendConfig:   true,
}

// IsEditableSetting reports whether the API may write a key.
//
// Membership in the table is not enough. A key is writable only if the
// application knows about it — config.SettingSpecs plus appDataSettings — so a
// row added to the database later does not become remotely writable by
// existing, and adding a sensitive one is a deliberate act rather than an
// oversight.
func IsEditableSetting(key string) bool {
	if appDataSettings[key] {
		return true
	}
	_, known := config.SettingSpecFor(key)
	return known
}

// SettingRequiresRestart reports whether a key is only read during startup, so
// the admin UI can say so instead of implying a change that has not happened.
func SettingRequiresRestart(key string) bool {
	spec, ok := config.SettingSpecFor(key)
	return ok && spec.RequiresRestart
}

// SettingEnvVar returns the environment variable a key overrides, or "" when
// the key is not configuration.
func SettingEnvVar(key string) string {
	spec, ok := config.SettingSpecFor(key)
	if !ok {
		return ""
	}
	return spec.EnvVar
}

// SettingsQuerier defines database operations for settings
type SettingsQuerier interface {
	GetAllSettings(ctx context.Context) ([]*sqlc.Setting, error)
	ProjectExists(ctx context.Context, projectID string) (bool, error)
	UpdateSettingText(ctx context.Context, arg sqlc.UpdateSettingTextParams) (int64, error)
	UpdateSettingInt(ctx context.Context, arg sqlc.UpdateSettingIntParams) (int64, error)
	UpdateSettingBool(ctx context.Context, arg sqlc.UpdateSettingBoolParams) (int64, error)
	UpdateSettingFloat(ctx context.Context, arg sqlc.UpdateSettingFloatParams) (int64, error)
	UpdateSettingJSON(ctx context.Context, arg sqlc.UpdateSettingJSONParams) (int64, error)
}

// SettingsTxRunner runs a function against a transactional querier, so a batch
// of settings either all land or none do.
type SettingsTxRunner interface {
	InTx(ctx context.Context, fn func(q SettingsQuerier) error) error
}

// SettingsService manages runtime configuration with in-memory caching
type SettingsService struct {
	queries     SettingsQuerier
	tx          SettingsTxRunner
	settingsMap atomic.Value // stores map[string]*sqlc.Setting
	logger      *slog.Logger
	overrideEnv bool
	stopRefresh chan struct{}
	refreshDone chan struct{}
}

// NewSettingsService creates a new settings service and starts background
// refresh. overrideEnv mirrors config.SettingsConfig.OverrideEnv: with it off,
// rows that shadow an environment variable are stored and served but never
// pushed into a running subsystem.
func NewSettingsService(ctx context.Context, queries SettingsQuerier, tx SettingsTxRunner, logger *slog.Logger, overrideEnv bool) (*SettingsService, error) {
	service := &SettingsService{
		queries:     queries,
		tx:          tx,
		logger:      logger,
		overrideEnv: overrideEnv,
		stopRefresh: make(chan struct{}),
		refreshDone: make(chan struct{}),
	}

	// Initial load is fail-fast: a misconfigured database must stop the server
	// at boot, where the cause is obvious, rather than later.
	if err := service.load(ctx); err != nil {
		return nil, fmt.Errorf("failed to load initial settings: %w", err)
	}

	// Start background refresh goroutine
	go service.backgroundRefresh()

	return service, nil
}

// backgroundRefresh periodically reloads settings from database
func (s *SettingsService) backgroundRefresh() {
	defer close(s.refreshDone)

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := s.RefreshSettings(ctx); err != nil {
				s.logger.Error("Failed to refresh settings", "error", err)
			}
			cancel()
		case <-s.stopRefresh:
			return
		}
	}
}

// RefreshSettings reloads all settings from the database.
//
// Unlike the initial load this never aborts the process. A bad value written
// directly to the database used to panic here — on the five-minute ticker, so
// the crash landed arbitrarily far from its cause. The previous map is kept
// instead, which is both valid and the value the server has been serving.
func (s *SettingsService) RefreshSettings(ctx context.Context) error {
	return s.load(ctx)
}

// load reads every setting, validates the result, and atomically swaps it in.
// The map is left untouched when the new one does not validate.
func (s *SettingsService) load(ctx context.Context) error {
	settings, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to query settings: %w", err)
	}

	newMap := make(map[string]*sqlc.Setting, len(settings))
	for _, setting := range settings {
		newMap[setting.Key] = setting
	}

	if err := s.validateSettings(ctx, newMap); err != nil {
		return fmt.Errorf("invalid setting in database: %w", err)
	}

	s.settingsMap.Store(newMap)
	s.applyLiveSettings(newMap)

	s.logger.Debug("Settings refreshed", "count", len(settings))
	return nil
}

// applyLiveSettings pushes the settings that can change without a restart into
// the subsystems that read them.
func (s *SettingsService) applyLiveSettings(settingsMap map[string]*sqlc.Setting) {
	if !s.overrideEnv {
		return
	}

	setting, ok := settingsMap["log_level"]
	if !ok || setting.ValueType != "text" || setting.ValueText == nil || *setting.ValueText == "" {
		return
	}

	if level := logger.ParseLevel(*setting.ValueText); level != logger.Level() {
		logger.SetLevel(level)
		s.logger.Info("Log level changed from settings", "level", *setting.ValueText)
	}
}

// validateSettings checks that critical settings are valid
func (s *SettingsService) validateSettings(ctx context.Context, settingsMap map[string]*sqlc.Setting) error {
	// Validate current_project_id exists in projects table
	setting, exists := settingsMap[SettingCurrentProjectID]
	if !exists {
		return fmt.Errorf("required setting '%s' not found in database", SettingCurrentProjectID)
	}

	if setting.ValueType != "text" || setting.ValueText == nil {
		return fmt.Errorf("setting '%s' must be of type 'text'", SettingCurrentProjectID)
	}

	projectID := *setting.ValueText
	projectExists, err := s.queries.ProjectExists(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to validate project existence: %w", err)
	}

	if !projectExists {
		return fmt.Errorf("%s '%s' does not exist in projects table", SettingCurrentProjectID, projectID)
	}

	return nil
}

// getSettingsMap returns the current settings map from atomic storage
func (s *SettingsService) getSettingsMap() map[string]*sqlc.Setting {
	value := s.settingsMap.Load()
	if value == nil {
		return make(map[string]*sqlc.Setting)
	}
	return value.(map[string]*sqlc.Setting)
}

// AllSettings returns every setting from the in-memory map, ordered by key.
func (s *SettingsService) AllSettings() []*sqlc.Setting {
	settingsMap := s.getSettingsMap()

	settings := make([]*sqlc.Setting, 0, len(settingsMap))
	for _, setting := range settingsMap {
		settings = append(settings, setting)
	}
	sort.Slice(settings, func(i, j int) bool { return settings[i].Key < settings[j].Key })

	return settings
}

// GetSetting returns a single setting from the in-memory map.
func (s *SettingsService) GetSetting(key string) (*sqlc.Setting, error) {
	setting, exists := s.getSettingsMap()[key]
	if !exists {
		return nil, fmt.Errorf("%w: %s", ErrSettingNotFound, key)
	}
	return setting, nil
}

// GetCurrentProjectID returns the default project ID for unauthenticated queries
func (s *SettingsService) GetCurrentProjectID(ctx context.Context) (string, error) {
	return s.GetTextSetting(ctx, SettingCurrentProjectID)
}

// SetCurrentProjectID points the whole system at a different project.
//
// The project is checked first and the write is skipped when it does not
// exist: an unknown ID would fail validation on the next load and leave the
// service serving a value the database no longer agrees with.
func (s *SettingsService) SetCurrentProjectID(ctx context.Context, projectID string) error {
	if err := s.validateProjectExists(ctx, projectID); err != nil {
		return err
	}

	rows, err := s.queries.UpdateSettingText(ctx, sqlc.UpdateSettingTextParams{
		Key:       SettingCurrentProjectID,
		ValueText: projectID,
	})
	if err != nil {
		return fmt.Errorf("failed to update %s: %w", SettingCurrentProjectID, err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: %s", ErrSettingNotFound, SettingCurrentProjectID)
	}

	return s.RefreshSettings(ctx)
}

// SettingUpdate is one key/value pair in a batch.
type SettingUpdate struct {
	Key   string
	Value string
}

// validateProjectExists guards the current-project write, which must never
// store an id that fails validation on the next load.
func (s *SettingsService) validateProjectExists(ctx context.Context, projectID string) error {
	exists, err := s.queries.ProjectExists(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to validate project existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: %s", ErrProjectNotFound, projectID)
	}
	return nil
}

// SetSettings writes a batch of existing keys.
//
// Every update is checked — known key, editable, value parses — before any of
// them is written, and the writes run in one transaction. A typo in the third
// field therefore leaves the first two untouched rather than half-applying the
// form.
func (s *SettingsService) SetSettings(ctx context.Context, updates []SettingUpdate) ([]*sqlc.Setting, error) {
	if len(updates) == 0 {
		return nil, nil
	}

	seen := make(map[string]bool, len(updates))
	for _, update := range updates {
		if seen[update.Key] {
			return nil, fmt.Errorf("%w: %s given twice", ErrInvalidSettingVal, update.Key)
		}
		seen[update.Key] = true

		setting, err := s.GetSetting(update.Key)
		if err != nil {
			return nil, err
		}
		if !IsEditableSetting(update.Key) {
			return nil, fmt.Errorf("%w: %s", ErrSettingNotEditable, update.Key)
		}
		if update.Key == SettingCurrentProjectID {
			if err := s.validateProjectExists(ctx, update.Value); err != nil {
				return nil, err
			}
			continue
		}
		if err := parseTyped(setting.ValueType, update.Value); err != nil {
			return nil, err
		}
	}

	err := s.tx.InTx(ctx, func(q SettingsQuerier) error {
		for _, update := range updates {
			setting, err := s.GetSetting(update.Key)
			if err != nil {
				return err
			}
			if err := writeTyped(ctx, q, update.Key, setting.ValueType, update.Value); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if err := s.RefreshSettings(ctx); err != nil {
		return nil, err
	}

	written := make([]*sqlc.Setting, 0, len(updates))
	for _, update := range updates {
		setting, err := s.GetSetting(update.Key)
		if err != nil {
			return nil, err
		}
		written = append(written, setting)
	}

	return written, nil
}

// parseTyped reports whether value is acceptable for valueType, without
// touching the database.
func parseTyped(valueType, value string) error {
	switch valueType {
	case "text":
		return nil
	case "int":
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return fmt.Errorf("%w: %q is not an integer", ErrInvalidSettingVal, value)
		}
	case "bool":
		if _, err := strconv.ParseBool(value); err != nil {
			return fmt.Errorf("%w: %q is not a boolean", ErrInvalidSettingVal, value)
		}
	case "float":
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return fmt.Errorf("%w: %q is not a number", ErrInvalidSettingVal, value)
		}
	case "json":
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("%w: not valid JSON", ErrInvalidSettingVal)
		}
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSettingType, valueType)
	}
	return nil
}

// writeTyped writes value to the column valueType names.
func writeTyped(ctx context.Context, q SettingsQuerier, key, valueType, value string) error {
	if err := parseTyped(valueType, value); err != nil {
		return err
	}

	var (
		rows int64
		err  error
	)

	switch valueType {
	case "text":
		rows, err = q.UpdateSettingText(ctx, sqlc.UpdateSettingTextParams{Key: key, ValueText: value})
	case "int":
		parsed, _ := strconv.ParseInt(value, 10, 64)
		rows, err = q.UpdateSettingInt(ctx, sqlc.UpdateSettingIntParams{Key: key, ValueInt: parsed})
	case "bool":
		parsed, _ := strconv.ParseBool(value)
		rows, err = q.UpdateSettingBool(ctx, sqlc.UpdateSettingBoolParams{Key: key, ValueBool: parsed})
	case "float":
		parsed, _ := strconv.ParseFloat(value, 64)
		rows, err = q.UpdateSettingFloat(ctx, sqlc.UpdateSettingFloatParams{Key: key, ValueFloat: parsed})
	case "json":
		rows, err = q.UpdateSettingJSON(ctx, sqlc.UpdateSettingJSONParams{Key: key, ValueJson: []byte(value)})
	}

	if err != nil {
		return fmt.Errorf("failed to update setting %s: %w", key, err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: %s", ErrSettingNotFound, key)
	}

	return nil
}

// SettingStringValue renders a setting's value in its canonical string form,
// which is what the GraphQL API exposes regardless of the column it lives in.
func SettingStringValue(setting *sqlc.Setting) string {
	switch setting.ValueType {
	case "text":
		if setting.ValueText != nil {
			return *setting.ValueText
		}
	case "int":
		if setting.ValueInt != nil {
			return strconv.FormatInt(*setting.ValueInt, 10)
		}
	case "bool":
		if setting.ValueBool != nil {
			return strconv.FormatBool(*setting.ValueBool)
		}
	case "float":
		if setting.ValueFloat != nil {
			return strconv.FormatFloat(*setting.ValueFloat, 'f', -1, 64)
		}
	case "json":
		if len(setting.ValueJson) > 0 {
			return string(setting.ValueJson)
		}
	}
	return ""
}

// GetTextSetting retrieves a text-typed setting from in-memory cache
func (s *SettingsService) GetTextSetting(ctx context.Context, key string) (string, error) {
	settingsMap := s.getSettingsMap()

	setting, exists := settingsMap[key]
	if !exists {
		return "", fmt.Errorf("%w: %s", ErrSettingNotFound, key)
	}

	if setting.ValueType != "text" {
		return "", fmt.Errorf("%w: expected text, got %s", ErrInvalidSettingType, setting.ValueType)
	}

	if setting.ValueText == nil {
		return "", fmt.Errorf("setting %s has null value", key)
	}

	return *setting.ValueText, nil
}

// GetIntSetting retrieves an int-typed setting from in-memory cache
func (s *SettingsService) GetIntSetting(ctx context.Context, key string) (int64, error) {
	settingsMap := s.getSettingsMap()

	setting, exists := settingsMap[key]
	if !exists {
		return 0, fmt.Errorf("%w: %s", ErrSettingNotFound, key)
	}

	if setting.ValueType != "int" {
		return 0, fmt.Errorf("%w: expected int, got %s", ErrInvalidSettingType, setting.ValueType)
	}

	if setting.ValueInt == nil {
		return 0, fmt.Errorf("setting %s has null value", key)
	}

	return *setting.ValueInt, nil
}

// GetBoolSetting retrieves a bool-typed setting from in-memory cache
func (s *SettingsService) GetBoolSetting(ctx context.Context, key string) (bool, error) {
	settingsMap := s.getSettingsMap()

	setting, exists := settingsMap[key]
	if !exists {
		return false, fmt.Errorf("%w: %s", ErrSettingNotFound, key)
	}

	if setting.ValueType != "bool" {
		return false, fmt.Errorf("%w: expected bool, got %s", ErrInvalidSettingType, setting.ValueType)
	}

	if setting.ValueBool == nil {
		return false, fmt.Errorf("setting %s has null value", key)
	}

	return *setting.ValueBool, nil
}

// GetFloatSetting retrieves a float-typed setting from in-memory cache
func (s *SettingsService) GetFloatSetting(ctx context.Context, key string) (float64, error) {
	settingsMap := s.getSettingsMap()

	setting, exists := settingsMap[key]
	if !exists {
		return 0, fmt.Errorf("%w: %s", ErrSettingNotFound, key)
	}

	if setting.ValueType != "float" {
		return 0, fmt.Errorf("%w: expected float, got %s", ErrInvalidSettingType, setting.ValueType)
	}

	if setting.ValueFloat == nil {
		return 0, fmt.Errorf("setting %s has null value", key)
	}

	return *setting.ValueFloat, nil
}

// Stop gracefully shuts down the background refresh goroutine
func (s *SettingsService) Stop() {
	close(s.stopRefresh)
	<-s.refreshDone
}

// pgxSettingsTx is the production SettingsTxRunner.
type pgxSettingsTx struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewSettingsTxRunner adapts a pool to SettingsTxRunner.
func NewSettingsTxRunner(pool *pgxpool.Pool, queries *sqlc.Queries) SettingsTxRunner {
	return &pgxSettingsTx{pool: pool, queries: queries}
}

func (s *pgxSettingsTx) InTx(ctx context.Context, fn func(q SettingsQuerier) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := fn(s.queries.WithTx(tx)); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
