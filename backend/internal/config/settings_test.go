package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func textValue(key, value string) SettingValue {
	return SettingValue{Key: key, ValueType: "text", Text: &value}
}

func boolValue(key string, value bool) SettingValue {
	return SettingValue{Key: key, ValueType: "bool", Bool: &value}
}

func floatValue(key string, value float64) SettingValue {
	return SettingValue{Key: key, ValueType: "float", Float: &value}
}

// envConfig stands in for what config.Load() produced from the environment.
func envConfig() *Config {
	cfg := &Config{}
	cfg.Log.Level = "info"
	cfg.Database.LogQueries = false
	cfg.OTEL.Enabled = false
	cfg.OTEL.SamplingRatio = 1.0
	cfg.SSF.DebugMode = false
	return cfg
}

func TestApplySettings_DatabaseWinsOverEnvironment(t *testing.T) {
	cfg := envConfig()

	applied := ApplySettings(cfg, map[string]SettingValue{
		"log_level":           textValue("log_level", "debug"),
		"db_log_queries":      boolValue("db_log_queries", true),
		"otel_enabled":        boolValue("otel_enabled", true),
		"otel_sampling_ratio": floatValue("otel_sampling_ratio", 0.25),
		"ssf_debug_mode":      boolValue("ssf_debug_mode", true),
	})

	assert.Equal(t, "debug", cfg.Log.Level)
	assert.True(t, cfg.Database.LogQueries)
	assert.True(t, cfg.OTEL.Enabled)
	assert.InDelta(t, 0.25, cfg.OTEL.SamplingRatio, 0)
	assert.True(t, cfg.SSF.DebugMode)

	assert.ElementsMatch(t, []string{
		"log_level", "db_log_queries", "otel_enabled",
		"otel_sampling_ratio", "ssf_debug_mode",
	}, applied)
}

// No rows at all is the fresh-database and unreachable-database case: the
// environment must survive untouched.
func TestApplySettings_WithoutRowsLeavesEnvironmentAlone(t *testing.T) {
	cfg := envConfig()

	applied := ApplySettings(cfg, nil)

	assert.Empty(t, applied)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.False(t, cfg.OTEL.Enabled)
	assert.InDelta(t, 1.0, cfg.OTEL.SamplingRatio, 0)
}

// "Use the row if it is defined" — a NULL or empty column is not defined, so
// the environment stands.
func TestApplySettings_SkipsUnsetRows(t *testing.T) {
	cfg := envConfig()

	applied := ApplySettings(cfg, map[string]SettingValue{
		"log_level":      {Key: "log_level", ValueType: "text"},
		"otel_enabled":   {Key: "otel_enabled", ValueType: "bool"},
		"ssf_debug_mode": boolValue("ssf_debug_mode", true),
	})

	assert.Equal(t, []string{"ssf_debug_mode"}, applied)
	assert.Equal(t, "info", cfg.Log.Level)
	assert.False(t, cfg.OTEL.Enabled)
	assert.True(t, cfg.SSF.DebugMode)
}

func TestApplySettings_EmptyTextIsNotDefined(t *testing.T) {
	cfg := envConfig()

	ApplySettings(cfg, map[string]SettingValue{
		"log_level": textValue("log_level", ""),
	})

	assert.Equal(t, "info", cfg.Log.Level)
}

// current_project_id is application data, not configuration; it must not reach
// any config field even though it rides in the same table.
func TestApplySettings_IgnoresNonConfigKeys(t *testing.T) {
	cfg := envConfig()

	applied := ApplySettings(cfg, map[string]SettingValue{
		"current_project_id": textValue("current_project_id", "PR01"),
	})

	assert.Empty(t, applied)
}

// A row whose value_type disagrees with its spec would make the apply
// function dereference a column that is nil.
func TestApplySettings_SurvivesAMistypedRow(t *testing.T) {
	cfg := envConfig()

	assert.NotPanics(t, func() {
		ApplySettings(cfg, map[string]SettingValue{
			"otel_enabled": {Key: "otel_enabled", ValueType: "nonsense", Bool: new(bool)},
		})
	})

	assert.False(t, cfg.OTEL.Enabled)
}

func TestSettingSpecFor(t *testing.T) {
	spec, ok := SettingSpecFor("otel_sampling_ratio")
	require.True(t, ok)
	assert.Equal(t, "OTEL_SAMPLING_RATIO", spec.EnvVar)
	assert.True(t, spec.RequiresRestart)

	// log_level is read through a slog.LevelVar, so it applies live.
	spec, ok = SettingSpecFor("log_level")
	require.True(t, ok)
	assert.False(t, spec.RequiresRestart)

	_, ok = SettingSpecFor("current_project_id")
	assert.False(t, ok)
}

func TestSettingValue_String(t *testing.T) {
	assert.Equal(t, "debug", textValue("k", "debug").String())
	assert.Equal(t, "true", boolValue("k", true).String())
	assert.Equal(t, "0.25", floatValue("k", 0.25).String())

	intVal := int64(42)
	assert.Equal(t, "42", SettingValue{ValueType: "int", Int: &intVal}.String())
	assert.Empty(t, SettingValue{ValueType: "text"}.String())
}

// Every spec must name a real environment variable, since the admin UI shows
// it as the fallback for that field.
func TestSettingSpecs_AllNameAnEnvVar(t *testing.T) {
	for _, spec := range SettingSpecs {
		assert.NotEmpty(t, spec.EnvVar, "spec %s", spec.Key)
		assert.NotNil(t, spec.apply, "spec %s", spec.Key)
	}
}
