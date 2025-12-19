-- +goose Up
CREATE TABLE user_feedback (
    id CHAR(28) PRIMARY KEY CHECK (id ~ '^FB[0-9A-Z]{26}$'),
    user_id CHAR(28) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    can_contact_me BOOLEAN NOT NULL DEFAULT FALSE,
    user_agent TEXT,
    platform TEXT,
    screen_width INT,
    screen_height INT,
    app_version TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_feedback_user_id ON user_feedback(user_id);
CREATE INDEX idx_user_feedback_created_at ON user_feedback(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS user_feedback;
