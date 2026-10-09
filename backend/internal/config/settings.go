package config

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// SettingValue is one row of the settings table. Not the sqlc model: this read
// happens before the pool those queries run on exists.
type SettingValue struct {
	Key       string
	ValueType string
	Text      *string
	Int       *int64
	Bool      *bool
	Float     *float64
}

// IsSet reports whether the row carries a value. An unset row falls back to
// the environment.
func (v SettingValue) IsSet() bool {
	switch v.ValueType {
	case "text":
		return v.Text != nil && *v.Text != ""
	case "int":
		return v.Int != nil
	case "bool":
		return v.Bool != nil
	case "float":
		return v.Float != nil
	}
	return false
}

// String renders the value the way the GraphQL API exposes it.
func (v SettingValue) String() string {
	switch v.ValueType {
	case "text":
		if v.Text != nil {
			return *v.Text
		}
	case "int":
		if v.Int != nil {
			return strconv.FormatInt(*v.Int, 10)
		}
	case "bool":
		if v.Bool != nil {
			return strconv.FormatBool(*v.Bool)
		}
	case "float":
		if v.Float != nil {
			return strconv.FormatFloat(*v.Float, 'f', -1, 64)
		}
	}
	return ""
}

// SettingSpec describes a settings row that backs a configuration field. The
// row wins over the environment when it is set.
type SettingSpec struct {
	Key    string
	EnvVar string
	// RequiresRestart is true when the field is only read during startup.
	RequiresRestart bool
	apply           func(cfg *Config, value SettingValue)
}

// SettingSpecs is the registry of settings that override configuration.
// `current_project_id` is absent on purpose: it is application data owned by
// services.SettingsService, not configuration.
var SettingSpecs = []SettingSpec{
	{
		Key:    "log_level",
		EnvVar: "LOG_LEVEL",
		// Read through a slog.LevelVar, so a refresh is enough.
		RequiresRestart: false,
		apply:           func(cfg *Config, v SettingValue) { cfg.Log.Level = *v.Text },
	},
	{
		Key:             "db_log_queries",
		EnvVar:          "DB_LOG_QUERIES",
		RequiresRestart: true,
		apply:           func(cfg *Config, v SettingValue) { cfg.Database.LogQueries = *v.Bool },
	},
	{
		Key:             "otel_enabled",
		EnvVar:          "OTEL_ENABLED",
		RequiresRestart: true,
		apply:           func(cfg *Config, v SettingValue) { cfg.OTEL.Enabled = *v.Bool },
	},
	{
		Key:             "otel_sampling_ratio",
		EnvVar:          "OTEL_SAMPLING_RATIO",
		RequiresRestart: true,
		apply:           func(cfg *Config, v SettingValue) { cfg.OTEL.SamplingRatio = *v.Float },
	},
	{
		Key:             "ssf_debug_mode",
		EnvVar:          "SSF_DEBUG_MODE",
		RequiresRestart: true,
		apply:           func(cfg *Config, v SettingValue) { cfg.SSF.DebugMode = *v.Bool },
	},
}

// SettingSpecFor returns the spec for a key, if the key backs a config field.
func SettingSpecFor(key string) (SettingSpec, bool) {
	for _, spec := range SettingSpecs {
		if spec.Key == key {
			return spec, true
		}
	}
	return SettingSpec{}, false
}

// ApplySettings overlays rows onto cfg and reports which keys it changed.
func ApplySettings(cfg *Config, rows map[string]SettingValue) []string {
	applied := make([]string, 0, len(SettingSpecs))

	for _, spec := range SettingSpecs {
		value, ok := rows[spec.Key]
		if !ok || !value.IsSet() {
			continue
		}
		// A mistyped row would make `apply` dereference a nil column.
		if _, err := validateType(spec, value); err != nil {
			slog.Warn("Ignoring settings row with unexpected type", "key", spec.Key, "error", err)
			continue
		}
		spec.apply(cfg, value)
		applied = append(applied, spec.Key)
	}

	return applied
}

func validateType(spec SettingSpec, value SettingValue) (SettingValue, error) {
	switch value.ValueType {
	case "text", "int", "bool", "float":
		return value, nil
	default:
		return value, fmt.Errorf("unsupported value_type %q for %s", value.ValueType, spec.Key)
	}
}

// LoadSettings reads the settings table over a connection of its own, because
// one of the values it resolves (db_log_queries) is an input to building the
// pool every other query runs on. Best-effort: an unreachable or unmigrated
// database leaves the environment in force.
func LoadSettings(ctx context.Context, databaseURL string) map[string]SettingValue {
	if databaseURL == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		slog.Warn("Could not read settings overrides, using environment only", "error", err)
		return nil
	}
	defer func() { _ = conn.Close(context.Background()) }()

	rows, err := conn.Query(ctx, `
		SELECT key, value_type, value_text, value_int, value_bool, value_float
		FROM settings
	`)
	if err != nil {
		// A fresh database has no settings table until migrations run later.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == undefinedTable {
			return nil
		}
		slog.Warn("Could not read settings overrides, using environment only", "error", err)
		return nil
	}
	defer rows.Close()

	values := make(map[string]SettingValue)
	for rows.Next() {
		var value SettingValue
		if err := rows.Scan(
			&value.Key, &value.ValueType,
			&value.Text, &value.Int, &value.Bool, &value.Float,
		); err != nil {
			slog.Warn("Could not read a settings row", "error", err)
			continue
		}
		values[value.Key] = value
	}
	if err := rows.Err(); err != nil {
		slog.Warn("Could not read settings overrides, using environment only", "error", err)
		return nil
	}

	return values
}

// undefinedTable is Postgres' SQLSTATE for a missing relation.
const undefinedTable = "42P01"
