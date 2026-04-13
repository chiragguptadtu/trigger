-- +goose Up
-- Revert 004: password_hash is required for email/password auth.
-- Rows inserted during the Google OAuth experiment (if any) get a sentinel empty hash.
UPDATE users SET password_hash = '' WHERE password_hash IS NULL;
ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash DROP DEFAULT;

-- +goose Down
ALTER TABLE users ALTER COLUMN password_hash DROP NOT NULL;
ALTER TABLE users ALTER COLUMN password_hash SET DEFAULT NULL;
