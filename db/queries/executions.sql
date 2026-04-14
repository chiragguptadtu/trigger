-- name: CreateExecution :one
INSERT INTO executions (command_id, triggered_by, inputs)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetExecutionByID :one
SELECT * FROM executions WHERE id = $1;

-- name: UpdateExecutionStatus :one
UPDATE executions
SET status        = $2,
    error_message = $3,
    started_at    = $4,
    completed_at  = $5
WHERE id = $1
RETURNING *;

-- name: ListExecutionsForCommand :many
SELECT e.*, u.name AS triggered_by_name, u.email AS triggered_by_email
FROM executions e
JOIN users u ON u.id = e.triggered_by
WHERE e.command_id = $1
ORDER BY e.created_at DESC
LIMIT $2;
