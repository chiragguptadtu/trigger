-- +goose Up
ALTER TABLE groups ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE groups DROP COLUMN is_active;
