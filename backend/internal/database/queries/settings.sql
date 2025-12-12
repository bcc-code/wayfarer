-- name: GetSettingByKey :one
SELECT key, value_text, value_int, value_bool, value_float, value_json, value_type, description, created_at, updated_at
FROM settings
WHERE key = @key::text;

-- name: GetSettingText :one
SELECT value_text
FROM settings
WHERE key = @key::text
  AND value_type = 'text';

-- name: GetSettingInt :one
SELECT value_int
FROM settings
WHERE key = @key::text
  AND value_type = 'int';

-- name: GetSettingBool :one
SELECT value_bool
FROM settings
WHERE key = @key::text
  AND value_type = 'bool';

-- name: GetSettingFloat :one
SELECT value_float
FROM settings
WHERE key = @key::text
  AND value_type = 'float';

-- name: GetSettingJSON :one
SELECT value_json
FROM settings
WHERE key = @key::text
  AND value_type = 'json';

-- name: GetAllSettings :many
SELECT key, value_text, value_int, value_bool, value_float, value_json, value_type, description, created_at, updated_at
FROM settings
ORDER BY key;

-- name: SetSettingText :exec
INSERT INTO settings (key, value_text, value_type, description)
VALUES (@key::text, @value_text::text, 'text', @description::text)
ON CONFLICT (key) DO UPDATE
    SET value_text = EXCLUDED.value_text,
        value_type = 'text',
        description = COALESCE(EXCLUDED.description, settings.description),
        updated_at = NOW();

-- name: SetSettingInt :exec
INSERT INTO settings (key, value_int, value_type, description)
VALUES (@key::text, @value_int::bigint, 'int', @description::text)
ON CONFLICT (key) DO UPDATE
    SET value_int = EXCLUDED.value_int,
        value_type = 'int',
        description = COALESCE(EXCLUDED.description, settings.description),
        updated_at = NOW();

-- name: SetSettingBool :exec
INSERT INTO settings (key, value_bool, value_type, description)
VALUES (@key::text, @value_bool::boolean, 'bool', @description::text)
ON CONFLICT (key) DO UPDATE
    SET value_bool = EXCLUDED.value_bool,
        value_type = 'bool',
        description = COALESCE(EXCLUDED.description, settings.description),
        updated_at = NOW();

-- name: SetSettingFloat :exec
INSERT INTO settings (key, value_float, value_type, description)
VALUES (@key::text, @value_float::double precision, 'float', @description::text)
ON CONFLICT (key) DO UPDATE
    SET value_float = EXCLUDED.value_float,
        value_type = 'float',
        description = COALESCE(EXCLUDED.description, settings.description),
        updated_at = NOW();

-- name: SetSettingJSON :exec
INSERT INTO settings (key, value_json, value_type, description)
VALUES (@key::text, @value_json::jsonb, 'json', @description::text)
ON CONFLICT (key) DO UPDATE
    SET value_json = EXCLUDED.value_json,
        value_type = 'json',
        description = COALESCE(EXCLUDED.description, settings.description),
        updated_at = NOW();

-- name: DeleteSetting :exec
DELETE FROM settings WHERE key = @key::text;
