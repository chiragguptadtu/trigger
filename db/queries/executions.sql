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
SELECT e.* FROM executions e
WHERE e.command_id = $1
  AND e.triggered_by = $2
ORDER BY e.created_at DESC
LIMIT $3;
