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

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

var (
	ErrSettingNotFound    = errors.New("setting not found")
	ErrInvalidSettingType = errors.New("invalid setting type")
	ErrInvalidSettingVal  = errors.New("invalid setting value")
	ErrProjectNotFound    = errors.New("project not found")
)

// SettingCurrentProjectID is the key holding the project every end user sees.
const SettingCurrentProjectID = "current_project_id"

// editableSettings are the keys an admin may change at runtime.
//
// Everything else in the table is inert: log_level, db_log_queries,
// otel_enabled, otel_sampling_ratio and ssf_debug_mode are all read from
// environment variables in internal/config, and no caller outside this file
// reads them through GetBoolSetting and friends. Offering them in the admin UI
// would be offering a knob that does nothing, so the API reports them as
// non-editable and refuses to write them.
var editableSettings = map[string]bool{
	SettingCurrentProjectID: true,
}

// IsEditableSetting reports whether a key may be changed at runtime.
func IsEditableSetting(key string) bool {
	return editableSettings[key]
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

// SettingsService manages runtime configuration with in-memory caching
type SettingsService struct {
	queries     SettingsQuerier
	settingsMap atomic.Value // stores map[string]*sqlc.Setting
	logger      *slog.Logger
	stopRefresh chan struct{}
	refreshDone chan struct{}
}

// NewSettingsService creates a new settings service and starts background refresh
func NewSettingsService(ctx context.Context, queries SettingsQuerier, logger *slog.Logger) (*SettingsService, error) {
	service := &SettingsService{
		queries:     queries,
		logger:      logger,
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

	s.logger.Debug("Settings refreshed", "count", len(settings))
	return nil
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
	exists, err := s.queries.ProjectExists(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to validate project existence: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: %s", ErrProjectNotFound, projectID)
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

// SetSetting writes the canonical string form of a value to an existing key.
//
// Only keys already in the table may be written — this is an editor, not a way
// to invent configuration — and only those on the editable allowlist, since
// every other row is read from the environment rather than from here.
func (s *SettingsService) SetSetting(ctx context.Context, key, value string) (*sqlc.Setting, error) {
	setting, err := s.GetSetting(key)
	if err != nil {
		return nil, err
	}

	if !IsEditableSetting(key) {
		return nil, fmt.Errorf("setting %s is not editable at runtime", key)
	}

	// Routed through SetCurrentProjectID so the existence check cannot be
	// bypassed by writing the same key through the generic path.
	if key == SettingCurrentProjectID {
		if err := s.SetCurrentProjectID(ctx, value); err != nil {
			return nil, err
		}
		return s.GetSetting(key)
	}

	if err := s.writeTyped(ctx, key, setting.ValueType, value); err != nil {
		return nil, err
	}

	if err := s.RefreshSettings(ctx); err != nil {
		return nil, err
	}

	return s.GetSetting(key)
}

// writeTyped parses value according to valueType and writes it to the matching
// column. A parse failure is reported before anything touches the database.
func (s *SettingsService) writeTyped(ctx context.Context, key, valueType, value string) error {
	var (
		rows int64
		err  error
	)

	switch valueType {
	case "text":
		rows, err = s.queries.UpdateSettingText(ctx, sqlc.UpdateSettingTextParams{Key: key, ValueText: value})
	case "int":
		parsed, parseErr := strconv.ParseInt(value, 10, 64)
		if parseErr != nil {
			return fmt.Errorf("%w: %q is not an integer", ErrInvalidSettingVal, value)
		}
		rows, err = s.queries.UpdateSettingInt(ctx, sqlc.UpdateSettingIntParams{Key: key, ValueInt: parsed})
	case "bool":
		parsed, parseErr := strconv.ParseBool(value)
		if parseErr != nil {
			return fmt.Errorf("%w: %q is not a boolean", ErrInvalidSettingVal, value)
		}
		rows, err = s.queries.UpdateSettingBool(ctx, sqlc.UpdateSettingBoolParams{Key: key, ValueBool: parsed})
	case "float":
		parsed, parseErr := strconv.ParseFloat(value, 64)
		if parseErr != nil {
			return fmt.Errorf("%w: %q is not a number", ErrInvalidSettingVal, value)
		}
		rows, err = s.queries.UpdateSettingFloat(ctx, sqlc.UpdateSettingFloatParams{Key: key, ValueFloat: parsed})
	case "json":
		if !json.Valid([]byte(value)) {
			return fmt.Errorf("%w: not valid JSON", ErrInvalidSettingVal)
		}
		rows, err = s.queries.UpdateSettingJSON(ctx, sqlc.UpdateSettingJSONParams{Key: key, ValueJson: []byte(value)})
	default:
		return fmt.Errorf("%w: %s", ErrInvalidSettingType, valueType)
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
