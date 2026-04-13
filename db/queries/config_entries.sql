-- name: CreateConfigEntry :one
INSERT INTO config_entries (key, value_encrypted, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetConfigEntryByKey :one
SELECT * FROM config_entries WHERE key = $1;

-- name: ListConfigEntries :many
SELECT id, key, description, created_by, created_at, updated_at
FROM config_entries
ORDER BY key;

-- name: UpdateConfigEntry :one
UPDATE config_entries
SET value_encrypted = $2, description = $3, updated_at = NOW()
WHERE key = $1
RETURNING *;

-- name: DeleteConfigEntry :exec
DELETE FROM config_entries WHERE key = $1;
