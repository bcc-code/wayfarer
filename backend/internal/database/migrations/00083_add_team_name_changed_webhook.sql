-- +goose Up
-- +goose StatementBegin

-- Add webhook event type for team name changed
ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS webhooks_event_type_check;
ALTER TABLE webhooks ADD CONSTRAINT webhooks_event_type_check
    CHECK (event_type IN ('external_content_event', 'points_awarded', 'quiz_session_finished', 'team_name_changed'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove team_name_changed from webhook event types
ALTER TABLE webhooks DROP CONSTRAINT IF EXISTS webhooks_event_type_check;
ALTER TABLE webhooks ADD CONSTRAINT webhooks_event_type_check
    CHECK (event_type IN ('external_content_event', 'points_awarded', 'quiz_session_finished'));

-- +goose StatementEnd
