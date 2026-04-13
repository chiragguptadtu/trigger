-- +goose Up
CREATE TABLE command_import_errors (
    filename  TEXT        PRIMARY KEY,
    error     TEXT        NOT NULL,
    failed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE command_import_errors;
