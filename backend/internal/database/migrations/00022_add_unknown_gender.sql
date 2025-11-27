-- +goose Up
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_gender_check;
ALTER TABLE users ADD CONSTRAINT users_gender_check CHECK (gender IN ('MALE', 'FEMALE', 'UNKNOWN'));

-- +goose Down
UPDATE users SET gender = 'MALE' WHERE gender = 'UNKNOWN';
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_gender_check;
ALTER TABLE users ADD CONSTRAINT users_gender_check CHECK (gender IN ('MALE', 'FEMALE'));
