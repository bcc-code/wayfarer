package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
)

var (
	ErrSettingNotFound    = errors.New("setting not found")
	ErrInvalidSettingType = errors.New("invalid setting type")
)

// SettingsQuerier defines database operations for settings
type SettingsQuerier interface {
	GetAllSettings(ctx context.Context) ([]*sqlc.Setting, error)
	ProjectExists(ctx context.Context, projectID string) (bool, error)
}

// SettingsService manages runtime configuration with in-memory caching
type SettingsService struct {
	queries     SettingsQuerier
	settingsMap atomic.Value // stores map[string]sqlc.Setting
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

	// Initial load of settings
	if err := service.RefreshSettings(ctx); err != nil {
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

// RefreshSettings reloads all settings from database and validates critical settings
func (s *SettingsService) RefreshSettings(ctx context.Context) error {
	// Load all settings from database
	settings, err := s.queries.GetAllSettings(ctx)
	if err != nil {
		return fmt.Errorf("failed to query settings: %w", err)
	}

	// Build new settings map
	newMap := make(map[string]*sqlc.Setting, len(settings))
	for _, setting := range settings {
		newMap[setting.Key] = setting
	}

	// Validate critical settings
	if err := s.validateSettings(ctx, newMap); err != nil {
		s.logger.Error("Invalid setting in database - crashing application", "error", err)
		panic(fmt.Sprintf("invalid setting in database: %v", err))
	}

	// Atomically swap the settings map
	s.settingsMap.Store(newMap)

	s.logger.Debug("Settings refreshed", "count", len(settings))
	return nil
}

// validateSettings checks that critical settings are valid
func (s *SettingsService) validateSettings(ctx context.Context, settingsMap map[string]*sqlc.Setting) error {
	// Validate current_project_id exists in projects table
	setting, exists := settingsMap["current_project_id"]
	if !exists {
		return fmt.Errorf("required setting 'current_project_id' not found in database")
	}

	if setting.ValueType != "text" || setting.ValueText == nil {
		return fmt.Errorf("setting 'current_project_id' must be of type 'text'")
	}

	projectID := *setting.ValueText
	projectExists, err := s.queries.ProjectExists(ctx, projectID)
	if err != nil {
		return fmt.Errorf("failed to validate project existence: %w", err)
	}

	if !projectExists {
		return fmt.Errorf("current_project_id '%s' does not exist in projects table", projectID)
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

// GetCurrentProjectID returns the default project ID for unauthenticated queries
func (s *SettingsService) GetCurrentProjectID(ctx context.Context) (string, error) {
	return s.GetTextSetting(ctx, "current_project_id")
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
