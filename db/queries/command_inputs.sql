-- name: CreateCommandInput :one
INSERT INTO command_inputs (command_id, name, label, type, options, multi, required, sort_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListCommandInputs :many
SELECT * FROM command_inputs
WHERE command_id = $1
ORDER BY sort_order, name;

-- name: DeleteCommandInputs :exec
DELETE FROM command_inputs WHERE command_id = $1;
