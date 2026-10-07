package services

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/bcc-media/wayfarer/internal/config"
	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/logger"
	"github.com/bcc-media/wayfarer/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	testProjectA = "PR01ARZ3NDEKTSV4RRFFQ69G5FA"
	testProjectB = "PR01ARZ3NDEKTSV4RRFFQ69G5FB"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func stringSetting(key, value string) *sqlc.Setting {
	return &sqlc.Setting{Key: key, ValueText: &value, ValueType: "text"}
}

func boolSetting(key string, value bool) *sqlc.Setting {
	return &sqlc.Setting{Key: key, ValueBool: &value, ValueType: "bool"}
}

// newServiceWithProject builds a service whose current_project_id is projectID.
// The background refresh goroutine is stopped by the test's cleanup, so the
// five-minute ticker never fires during a test.
func newServiceWithProject(t *testing.T, queries *mocks.MockSettingsQuerier, projectID string, extra ...*sqlc.Setting) *SettingsService {
	t.Helper()

	settings := append([]*sqlc.Setting{stringSetting(SettingCurrentProjectID, projectID)}, extra...)
	queries.On("GetAllSettings", mock.Anything).Return(settings, nil).Once()
	queries.On("ProjectExists", mock.Anything, projectID).Return(true, nil).Once()

	service, err := NewSettingsService(context.Background(), queries, testLogger())
	require.NoError(t, err)
	t.Cleanup(service.Stop)

	return service
}

func TestNewSettingsService_LoadsCurrentProject(t *testing.T) {
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	projectID, err := service.GetCurrentProjectID(context.Background())
	require.NoError(t, err)
	assert.Equal(t, testProjectA, projectID)
}

func TestNewSettingsService_FailsFastOnMissingCurrentProject(t *testing.T) {
	queries := mocks.NewMockSettingsQuerier(t)
	queries.On("GetAllSettings", mock.Anything).Return([]*sqlc.Setting{}, nil).Once()

	service, err := NewSettingsService(context.Background(), queries, testLogger())

	require.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), SettingCurrentProjectID)
}

func TestNewSettingsService_FailsFastWhenProjectDoesNotExist(t *testing.T) {
	queries := mocks.NewMockSettingsQuerier(t)
	queries.On("GetAllSettings", mock.Anything).
		Return([]*sqlc.Setting{stringSetting(SettingCurrentProjectID, testProjectA)}, nil).Once()
	queries.On("ProjectExists", mock.Anything, testProjectA).Return(false, nil).Once()

	service, err := NewSettingsService(context.Background(), queries, testLogger())

	require.Error(t, err)
	assert.Nil(t, service)
}

// A bad value written directly to the database used to panic here, taking the
// process down from the five-minute background ticker — arbitrarily far from
// whatever caused it. The previous map must survive instead.
func TestRefreshSettings_KeepsPreviousMapWhenNewOneIsInvalid(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	queries.On("GetAllSettings", mock.Anything).
		Return([]*sqlc.Setting{stringSetting(SettingCurrentProjectID, "PR0000000000000000000000000")}, nil).Once()
	queries.On("ProjectExists", mock.Anything, "PR0000000000000000000000000").Return(false, nil).Once()

	assert.NotPanics(t, func() {
		err := service.RefreshSettings(ctx)
		assert.Error(t, err)
	})

	projectID, err := service.GetCurrentProjectID(ctx)
	require.NoError(t, err)
	assert.Equal(t, testProjectA, projectID, "the previously valid project must still be served")
}

func TestRefreshSettings_KeepsPreviousMapOnQueryError(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	queries.On("GetAllSettings", mock.Anything).Return(nil, errors.New("connection refused")).Once()

	require.Error(t, service.RefreshSettings(ctx))

	projectID, err := service.GetCurrentProjectID(ctx)
	require.NoError(t, err)
	assert.Equal(t, testProjectA, projectID)
}

func TestSetCurrentProjectID_WritesAndRefreshes(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	queries.On("ProjectExists", mock.Anything, testProjectB).Return(true, nil).Once()
	queries.On("UpdateSettingText", mock.Anything, sqlc.UpdateSettingTextParams{
		Key:       SettingCurrentProjectID,
		ValueText: testProjectB,
	}).Return(int64(1), nil).Once()

	// The refresh that follows the write.
	queries.On("GetAllSettings", mock.Anything).
		Return([]*sqlc.Setting{stringSetting(SettingCurrentProjectID, testProjectB)}, nil).Once()
	queries.On("ProjectExists", mock.Anything, testProjectB).Return(true, nil).Once()

	require.NoError(t, service.SetCurrentProjectID(ctx, testProjectB))

	projectID, err := service.GetCurrentProjectID(ctx)
	require.NoError(t, err)
	assert.Equal(t, testProjectB, projectID)
}

// The existence check must come before the write: an unknown ID would
// otherwise be persisted and then fail validation on the next load.
func TestSetCurrentProjectID_RejectsUnknownProjectBeforeWriting(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	queries.On("ProjectExists", mock.Anything, testProjectB).Return(false, nil).Once()

	err := service.SetCurrentProjectID(ctx, testProjectB)

	require.ErrorIs(t, err, ErrProjectNotFound)
	queries.AssertNotCalled(t, "UpdateSettingText", mock.Anything, mock.Anything)

	projectID, err := service.GetCurrentProjectID(ctx)
	require.NoError(t, err)
	assert.Equal(t, testProjectA, projectID)
}

func TestSetCurrentProjectID_ErrorsWhenRowIsMissing(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	queries.On("ProjectExists", mock.Anything, testProjectB).Return(true, nil).Once()
	queries.On("UpdateSettingText", mock.Anything, mock.Anything).Return(int64(0), nil).Once()

	err := service.SetCurrentProjectID(ctx, testProjectB)

	require.ErrorIs(t, err, ErrSettingNotFound)
}

func TestSetSetting_RoutesCurrentProjectThroughTheExistenceCheck(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	queries.On("ProjectExists", mock.Anything, testProjectB).Return(false, nil).Once()

	setting, err := service.SetSetting(ctx, SettingCurrentProjectID, testProjectB)

	require.ErrorIs(t, err, ErrProjectNotFound)
	assert.Nil(t, setting)
	queries.AssertNotCalled(t, "UpdateSettingText", mock.Anything, mock.Anything)
}

func TestSetSetting_RejectsUnknownKey(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA)

	setting, err := service.SetSetting(ctx, "not_a_real_key", "whatever")

	require.ErrorIs(t, err, ErrSettingNotFound)
	assert.Nil(t, setting)
}

// Being a row is not enough to be writable. Without this, adding a settings
// row later would silently make it remotely writable.
func TestSetSetting_RejectsAKeyTheApplicationDoesNotKnow(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA,
		stringSetting("some_future_api_key", "s3cret"))

	setting, err := service.SetSetting(ctx, "some_future_api_key", "attacker-controlled")

	require.ErrorIs(t, err, ErrSettingNotEditable)
	assert.Nil(t, setting)
	queries.AssertNotCalled(t, "UpdateSettingText", mock.Anything, mock.Anything)
}

func TestIsEditableSetting(t *testing.T) {
	assert.True(t, IsEditableSetting(SettingCurrentProjectID))
	for _, spec := range config.SettingSpecs {
		assert.True(t, IsEditableSetting(spec.Key), spec.Key)
	}
	assert.False(t, IsEditableSetting("frontend_config"))
	assert.False(t, IsEditableSetting("anything_else"))
}

func TestSetSetting_WritesANonProjectKey(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA, boolSetting("otel_enabled", true))

	queries.On("UpdateSettingBool", mock.Anything, sqlc.UpdateSettingBoolParams{
		Key:       "otel_enabled",
		ValueBool: false,
	}).Return(int64(1), nil).Once()

	// The refresh that follows the write.
	queries.On("GetAllSettings", mock.Anything).Return([]*sqlc.Setting{
		stringSetting(SettingCurrentProjectID, testProjectA),
		boolSetting("otel_enabled", false),
	}, nil).Once()
	queries.On("ProjectExists", mock.Anything, testProjectA).Return(true, nil).Once()

	setting, err := service.SetSetting(ctx, "otel_enabled", "false")

	require.NoError(t, err)
	assert.Equal(t, "false", SettingStringValue(setting))
}

// log_level is read through a slog.LevelVar, so a refresh changes verbosity
// without a restart. Everything else is only read at startup.
func TestRefreshSettings_AppliesLogLevelLive(t *testing.T) {
	ctx := context.Background()
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA, stringSetting("log_level", "info"))
	t.Cleanup(func() { logger.SetLevel(slog.LevelInfo) })

	require.Equal(t, slog.LevelInfo, logger.Level())

	queries.On("GetAllSettings", mock.Anything).Return([]*sqlc.Setting{
		stringSetting(SettingCurrentProjectID, testProjectA),
		stringSetting("log_level", "debug"),
	}, nil).Once()
	queries.On("ProjectExists", mock.Anything, testProjectA).Return(true, nil).Once()

	require.NoError(t, service.RefreshSettings(ctx))

	assert.Equal(t, slog.LevelDebug, logger.Level())
}

func TestWriteTyped_ParsesEachValueType(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		valueType string
		value     string
		method    string
		expectArg any
	}{
		{"text", "text", "hello", "UpdateSettingText", sqlc.UpdateSettingTextParams{Key: "k", ValueText: "hello"}},
		{"int", "int", "42", "UpdateSettingInt", sqlc.UpdateSettingIntParams{Key: "k", ValueInt: 42}},
		{"bool", "bool", "true", "UpdateSettingBool", sqlc.UpdateSettingBoolParams{Key: "k", ValueBool: true}},
		{"float", "float", "0.25", "UpdateSettingFloat", sqlc.UpdateSettingFloatParams{Key: "k", ValueFloat: 0.25}},
		{"json", "json", `{"a":1}`, "UpdateSettingJSON", sqlc.UpdateSettingJSONParams{Key: "k", ValueJson: []byte(`{"a":1}`)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := mocks.NewMockSettingsQuerier(t)
			service := newServiceWithProject(t, queries, testProjectA)

			queries.On(tt.method, mock.Anything, tt.expectArg).Return(int64(1), nil).Once()

			require.NoError(t, service.writeTyped(ctx, "k", tt.valueType, tt.value))
		})
	}
}

func TestWriteTyped_RejectsMalformedValues(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		valueType string
		value     string
	}{
		{"int", "int", "twelve"},
		{"bool", "bool", "yes please"},
		{"float", "float", "π"},
		{"json", "json", "{not json"},
		{"unknown type", "duration", "5m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries := mocks.NewMockSettingsQuerier(t)
			service := newServiceWithProject(t, queries, testProjectA)

			err := service.writeTyped(ctx, "k", tt.valueType, tt.value)

			require.Error(t, err)
			// Nothing may reach the database when the value does not parse.
			queries.AssertNotCalled(t, "UpdateSettingInt", mock.Anything, mock.Anything)
			queries.AssertNotCalled(t, "UpdateSettingBool", mock.Anything, mock.Anything)
			queries.AssertNotCalled(t, "UpdateSettingFloat", mock.Anything, mock.Anything)
			queries.AssertNotCalled(t, "UpdateSettingJSON", mock.Anything, mock.Anything)
		})
	}
}

func TestAllSettings_IsOrderedByKey(t *testing.T) {
	queries := mocks.NewMockSettingsQuerier(t)
	service := newServiceWithProject(t, queries, testProjectA,
		stringSetting("log_level", "info"),
		boolSetting("otel_enabled", true),
	)

	settings := service.AllSettings()

	keys := make([]string, 0, len(settings))
	for _, setting := range settings {
		keys = append(keys, setting.Key)
	}
	assert.Equal(t, []string{SettingCurrentProjectID, "log_level", "otel_enabled"}, keys)
}

func TestSettingStringValue(t *testing.T) {
	intValue := int64(42)
	floatValue := 0.25
	boolValue := false

	tests := []struct {
		name    string
		setting *sqlc.Setting
		want    string
	}{
		{"text", stringSetting("k", "hello"), "hello"},
		{"int", &sqlc.Setting{ValueType: "int", ValueInt: &intValue}, "42"},
		{"bool", &sqlc.Setting{ValueType: "bool", ValueBool: &boolValue}, "false"},
		{"float", &sqlc.Setting{ValueType: "float", ValueFloat: &floatValue}, "0.25"},
		{"json", &sqlc.Setting{ValueType: "json", ValueJson: []byte(`{"a":1}`)}, `{"a":1}`},
		{"null column", &sqlc.Setting{ValueType: "text"}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SettingStringValue(tt.setting))
		})
	}
}

// Which keys need a restart is what the admin UI shows next to each field, so
// a key moving between the two groups should be a deliberate edit.
func TestSettingRequiresRestart(t *testing.T) {
	// Read through a LevelVar, so it applies on the next refresh.
	assert.False(t, SettingRequiresRestart("log_level"))
	// Read once, while the process starts up.
	assert.True(t, SettingRequiresRestart("otel_enabled"))
	assert.True(t, SettingRequiresRestart("otel_sampling_ratio"))
	assert.True(t, SettingRequiresRestart("db_log_queries"))
	assert.True(t, SettingRequiresRestart("ssf_debug_mode"))
	// Application data, not configuration.
	assert.False(t, SettingRequiresRestart(SettingCurrentProjectID))
}

func TestSettingEnvVar(t *testing.T) {
	assert.Equal(t, "LOG_LEVEL", SettingEnvVar("log_level"))
	assert.Equal(t, "OTEL_SAMPLING_RATIO", SettingEnvVar("otel_sampling_ratio"))
	assert.Empty(t, SettingEnvVar(SettingCurrentProjectID))
}
