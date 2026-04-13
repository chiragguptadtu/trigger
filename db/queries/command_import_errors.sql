-- name: UpsertImportError :exec
INSERT INTO command_import_errors (filename, error, failed_at)
VALUES ($1, $2, NOW())
ON CONFLICT (filename) DO UPDATE
    SET error     = EXCLUDED.error,
        failed_at = EXCLUDED.failed_at;

-- name: DeleteImportError :exec
DELETE FROM command_import_errors WHERE filename = $1;

-- name: ClearImportErrors :exec
DELETE FROM command_import_errors;

-- name: ListImportErrors :many
SELECT filename, error, failed_at
FROM command_import_errors
ORDER BY filename;
