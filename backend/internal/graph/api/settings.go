package api

import (
	"context"
	"fmt"

	"github.com/bcc-media/wayfarer/internal/database/sqlc"
	"github.com/bcc-media/wayfarer/internal/graph/api/model"
	"github.com/bcc-media/wayfarer/internal/graph/scalars"
	"github.com/bcc-media/wayfarer/internal/middleware"
	"github.com/bcc-media/wayfarer/internal/services"
)

// settingValueTypes maps the database's value_type strings to the enum.
var settingValueTypes = map[string]model.SettingValueType{
	"text":  model.SettingValueTypeText,
	"int":   model.SettingValueTypeInt,
	"bool":  model.SettingValueTypeBool,
	"float": model.SettingValueTypeFloat,
	"json":  model.SettingValueTypeJSON,
}

// sqlcSettingToModel converts a settings row to its GraphQL representation,
// collapsing the five value columns into the one the row's type names.
func sqlcSettingToModel(setting *sqlc.Setting) model.Setting {
	valueType, ok := settingValueTypes[setting.ValueType]
	if !ok {
		// CHECK-constrained in the schema, so this is unreachable short of a
		// migration adding a type without updating the map.
		valueType = model.SettingValueTypeText
	}

	converted := model.Setting{
		Key:             setting.Key,
		Value:           services.SettingStringValue(setting),
		ValueType:       valueType,
		Description:     setting.Description,
		RequiresRestart: services.SettingRequiresRestart(setting.Key),
		Editable:        services.IsEditableSetting(setting.Key),
	}
	if envVar := services.SettingEnvVar(setting.Key); envVar != "" {
		converted.EnvVar = &envVar
	}
	if updatedAt := scalars.ToDateTimePointer(setting.UpdatedAt); updatedAt != nil {
		converted.UpdatedAt = *updatedAt
	}

	return converted
}

// requireSuperAdmin is the in-resolver authorization used by queries, which
// carry no @requireRole directive.
func (r *Resolver) requireSuperAdmin(ctx context.Context) error {
	userID, ok := middleware.GetUserID(ctx)
	if !ok || userID == "" {
		return fmt.Errorf("user not authenticated")
	}
	if !r.RoleService.IsSuperAdmin(ctx, userID) {
		return fmt.Errorf("permission denied: superadmin role required")
	}
	return nil
}

// listSettings returns every setting, ordered by key.
func (r *Resolver) listSettings(ctx context.Context) ([]model.Setting, error) {
	if err := r.requireSuperAdmin(ctx); err != nil {
		return nil, err
	}

	rows := r.SettingsService.AllSettings()
	settings := make([]model.Setting, 0, len(rows))
	for _, row := range rows {
		settings = append(settings, sqlcSettingToModel(row))
	}

	return settings, nil
}

// setCurrentProject points the whole system at another project.
func (r *Resolver) setCurrentProject(ctx context.Context, projectID string) (*model.Project, error) {
	// Read the outgoing project before the write so both can be invalidated;
	// a missing one is not fatal, there is simply nothing to invalidate.
	previousProjectID, err := r.SettingsService.GetCurrentProjectID(ctx)
	if err != nil {
		previousProjectID = ""
	}

	if err := r.SettingsService.SetCurrentProjectID(ctx, projectID); err != nil {
		return nil, fmt.Errorf("failed to set current project: %w", err)
	}

	r.invalidateCurrentProject(previousProjectID, projectID)

	return r.LoadProjectWithTranslation(ctx, projectID)
}

// setSetting writes one setting and returns the stored result.
func (r *Resolver) setSetting(ctx context.Context, key, value string) (*model.Setting, error) {
	previousProjectID := ""
	if key == services.SettingCurrentProjectID {
		if current, err := r.SettingsService.GetCurrentProjectID(ctx); err == nil {
			previousProjectID = current
		}
	}

	setting, err := r.SettingsService.SetSetting(ctx, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to set setting %s: %w", key, err)
	}

	if key == services.SettingCurrentProjectID {
		r.invalidateCurrentProject(previousProjectID, value)
	} else {
		r.Cache.InvalidateSettings()
	}

	converted := sqlcSettingToModel(setting)
	return &converted, nil
}

// invalidateCurrentProject clears what a change of current project makes stale,
// on this instance and on every other one.
//
// InvalidateProject drops the whole gqlresponse: prefix, which is what evicts
// the cached currentProject / myCurrentProject responses — without it the old
// project keeps being served until the response cache's TTL expires. Both the
// outgoing and incoming projects are invalidated because entries exist for
// each. InvalidateSettings is separate: it reloads the settings map itself,
// which no cache invalidation would otherwise touch.
func (r *Resolver) invalidateCurrentProject(previousProjectID, projectID string) {
	if previousProjectID != "" && previousProjectID != projectID {
		r.Cache.InvalidateProject(previousProjectID)
	}
	r.Cache.InvalidateProject(projectID)
	r.Cache.InvalidateSettings()
}
