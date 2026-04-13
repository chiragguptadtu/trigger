-- name: UpsertCommand :one
INSERT INTO commands (slug, name, description, script_path)
VALUES ($1, $2, $3, $4)
ON CONFLICT (slug) DO UPDATE
    SET name        = EXCLUDED.name,
        description = EXCLUDED.description,
        script_path = EXCLUDED.script_path,
        is_active   = true
RETURNING *;

-- name: DeactivateCommand :exec
UPDATE commands SET is_active = false WHERE slug = $1;

-- name: GetCommandBySlug :one
SELECT * FROM commands WHERE slug = $1 AND is_active = true;

-- name: GetCommandByID :one
SELECT * FROM commands WHERE id = $1;

-- name: ListAllCommands :many
SELECT * FROM commands WHERE is_active = true ORDER BY name;

-- name: ListCommandsForUser :many
SELECT DISTINCT c.* FROM commands c
LEFT JOIN command_permissions cp ON cp.command_id = c.id
LEFT JOIN user_group_members ugm ON ugm.user_id = $1 AND ugm.group_id = cp.grantee_id
WHERE c.is_active = true
  AND (
      cp.grantee_type = 'user' AND cp.grantee_id = $1
   OR cp.grantee_type = 'group' AND ugm.user_id = $1
  )
ORDER BY c.name;
