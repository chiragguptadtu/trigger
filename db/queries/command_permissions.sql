-- name: CreateCommandPermission :one
INSERT INTO command_permissions (command_id, grantee_type, grantee_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING
RETURNING *;

-- name: DeleteCommandPermission :exec
DELETE FROM command_permissions
WHERE command_id = $1 AND grantee_type = $2 AND grantee_id = $3;

-- name: ListCommandPermissions :many
SELECT * FROM command_permissions WHERE command_id = $1;

-- name: DeleteAllCommandPermissions :exec
DELETE FROM command_permissions WHERE command_id = $1;

-- name: UserCanAccessCommand :one
SELECT EXISTS (
    SELECT 1 FROM command_permissions cp
    LEFT JOIN user_group_members ugm
           ON ugm.user_id = $1 AND ugm.group_id = cp.grantee_id
    WHERE cp.command_id = $2
      AND (
          (cp.grantee_type = 'user'  AND cp.grantee_id = $1)
       OR (cp.grantee_type = 'group' AND ugm.user_id   IS NOT NULL)
      )
) AS has_access;
