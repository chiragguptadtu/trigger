-- name: CreateGroup :one
INSERT INTO groups (name) VALUES ($1) RETURNING *;

-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = $1;

-- name: ListGroups :many
SELECT * FROM groups WHERE is_active = true ORDER BY name;

-- name: DeactivateGroup :exec
UPDATE groups SET is_active = false WHERE id = $1;

-- name: AddGroupMember :exec
INSERT INTO user_group_members (user_id, group_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveGroupMember :exec
DELETE FROM user_group_members WHERE user_id = $1 AND group_id = $2;

-- name: ListGroupMembers :many
SELECT u.* FROM users u
JOIN user_group_members m ON m.user_id = u.id
WHERE m.group_id = $1
ORDER BY u.name;

-- name: ListUserGroups :many
SELECT g.* FROM groups g
JOIN user_group_members m ON m.group_id = g.id
WHERE m.user_id = $1
ORDER BY g.name;
