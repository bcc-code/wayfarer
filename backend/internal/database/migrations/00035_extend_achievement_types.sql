-- +goose Up
-- +goose StatementBegin

-- Migration: Extend achievement types to include QUIZ
-- This migration adds the QUIZ achievement type to support quiz-based achievements

-- Drop the old CHECK constraint
ALTER TABLE achievements DROP CONSTRAINT IF EXISTS achievements_achievement_type_check;

-- Add the new CHECK constraint with QUIZ type
ALTER TABLE achievements ADD CONSTRAINT achievements_achievement_type_check
    CHECK (achievement_type IN ('SIMPLE', 'READING', 'LISTENING', 'STREAK', 'QUIZ'));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Revert to original CHECK constraint without QUIZ
ALTER TABLE achievements DROP CONSTRAINT IF EXISTS achievements_achievement_type_check;
ALTER TABLE achievements ADD CONSTRAINT achievements_achievement_type_check
    CHECK (achievement_type IN ('SIMPLE', 'READING', 'LISTENING', 'STREAK'));

-- +goose StatementEnd
