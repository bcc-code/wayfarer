-- +goose Up
-- +goose StatementBegin

-- Add new light mode color columns
ALTER TABLE projects ADD COLUMN color_light_accent VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_accent_contrast VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_on_accent VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_background_default VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_background_raised VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_background_indent VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_text_default VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_text_muted VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_text_hint VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_shadow_default VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_shadow_blank VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_light_border_default VARCHAR(50);

-- Add new dark mode color columns
ALTER TABLE projects ADD COLUMN color_dark_accent VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_accent_contrast VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_on_accent VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_background_default VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_background_raised VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_background_indent VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_text_default VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_text_muted VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_text_hint VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_shadow_default VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_shadow_blank VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_dark_border_default VARCHAR(50);

-- Set default values for existing rows
UPDATE projects SET
    -- Light mode defaults
    color_light_accent = '#e8dfa7',
    color_light_accent_contrast = '#938636',
    color_light_on_accent = '#01121a',
    color_light_background_default = '#f3ede5',
    color_light_background_raised = '#ffffff',
    color_light_background_indent = 'rgb(99 56 1 / 0.05)',
    color_light_text_default = '#282521',
    color_light_text_muted = 'rgb(40 37 33 / 0.65)',
    color_light_text_hint = 'rgb(40 37 33 / 0.4)',
    color_light_shadow_default = 'rgb(40 37 33 / 0.1)',
    color_light_shadow_blank = 'rgb(40 37 33 / 0)',
    color_light_border_default = 'rgb(40 37 33 / 0.15)',
    -- Dark mode defaults
    color_dark_accent = '#e8dfa7',
    color_dark_accent_contrast = '#e8dfa7',
    color_dark_on_accent = '#1a1401',
    color_dark_background_default = '#122026',
    color_dark_background_raised = '#0a3644',
    color_dark_background_indent = 'rgb(0 9 13 / 0.25)',
    color_dark_text_default = '#f3ede5',
    color_dark_text_muted = 'rgb(243 237 229 / 0.7)',
    color_dark_text_hint = 'rgb(243 237 229 / 0.4)',
    color_dark_shadow_default = 'rgb(18 32 38 / 0.3)',
    color_dark_shadow_blank = 'rgb(18 32 38 / 0)',
    color_dark_border_default = 'rgb(156 214 243 / 0.09)';

-- Make columns NOT NULL after setting defaults
ALTER TABLE projects ALTER COLUMN color_light_accent SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_accent_contrast SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_on_accent SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_background_default SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_background_raised SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_background_indent SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_text_default SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_text_muted SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_text_hint SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_shadow_default SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_shadow_blank SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_light_border_default SET NOT NULL;

ALTER TABLE projects ALTER COLUMN color_dark_accent SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_accent_contrast SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_on_accent SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_background_default SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_background_raised SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_background_indent SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_text_default SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_text_muted SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_text_hint SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_shadow_default SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_shadow_blank SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_dark_border_default SET NOT NULL;

-- Drop old color columns
ALTER TABLE projects DROP COLUMN color_primary;
ALTER TABLE projects DROP COLUMN color_secondary;
ALTER TABLE projects DROP COLUMN color_tertiary;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Re-add old color columns
ALTER TABLE projects ADD COLUMN color_primary VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_secondary VARCHAR(50);
ALTER TABLE projects ADD COLUMN color_tertiary VARCHAR(50);

-- Set defaults for old columns from light mode accent colors
UPDATE projects SET
    color_primary = color_light_accent,
    color_secondary = color_light_accent_contrast,
    color_tertiary = color_light_on_accent;

ALTER TABLE projects ALTER COLUMN color_primary SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_secondary SET NOT NULL;
ALTER TABLE projects ALTER COLUMN color_tertiary SET NOT NULL;

-- Drop new color columns
ALTER TABLE projects DROP COLUMN color_light_accent;
ALTER TABLE projects DROP COLUMN color_light_accent_contrast;
ALTER TABLE projects DROP COLUMN color_light_on_accent;
ALTER TABLE projects DROP COLUMN color_light_background_default;
ALTER TABLE projects DROP COLUMN color_light_background_raised;
ALTER TABLE projects DROP COLUMN color_light_background_indent;
ALTER TABLE projects DROP COLUMN color_light_text_default;
ALTER TABLE projects DROP COLUMN color_light_text_muted;
ALTER TABLE projects DROP COLUMN color_light_text_hint;
ALTER TABLE projects DROP COLUMN color_light_shadow_default;
ALTER TABLE projects DROP COLUMN color_light_shadow_blank;
ALTER TABLE projects DROP COLUMN color_light_border_default;

ALTER TABLE projects DROP COLUMN color_dark_accent;
ALTER TABLE projects DROP COLUMN color_dark_accent_contrast;
ALTER TABLE projects DROP COLUMN color_dark_on_accent;
ALTER TABLE projects DROP COLUMN color_dark_background_default;
ALTER TABLE projects DROP COLUMN color_dark_background_raised;
ALTER TABLE projects DROP COLUMN color_dark_background_indent;
ALTER TABLE projects DROP COLUMN color_dark_text_default;
ALTER TABLE projects DROP COLUMN color_dark_text_muted;
ALTER TABLE projects DROP COLUMN color_dark_text_hint;
ALTER TABLE projects DROP COLUMN color_dark_shadow_default;
ALTER TABLE projects DROP COLUMN color_dark_shadow_blank;
ALTER TABLE projects DROP COLUMN color_dark_border_default;

-- +goose StatementEnd
