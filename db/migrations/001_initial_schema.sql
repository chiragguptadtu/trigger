-- +goose Up

CREATE TABLE users (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT        NOT NULL UNIQUE,
    name        TEXT        NOT NULL,
    password_hash TEXT      NOT NULL,
    is_admin    BOOLEAN     NOT NULL DEFAULT false,
    is_active   BOOLEAN     NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE groups (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE user_group_members (
    user_id  UUID NOT NULL REFERENCES users(id)  ON DELETE CASCADE,
    group_id UUID NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, group_id)
);

CREATE TABLE commands (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    slug          TEXT        NOT NULL UNIQUE,
    name          TEXT        NOT NULL,
    description   TEXT        NOT NULL DEFAULT '',
    script_path   TEXT        NOT NULL,
    is_active     BOOLEAN     NOT NULL DEFAULT true,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE command_inputs (
    id         UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
    command_id UUID    NOT NULL REFERENCES commands(id) ON DELETE CASCADE,
    name       TEXT    NOT NULL,
    label      TEXT    NOT NULL,
    type       TEXT    NOT NULL CHECK (type IN ('open', 'closed')),
    options    JSONB,
    multi      BOOLEAN NOT NULL DEFAULT false,
    required   BOOLEAN NOT NULL DEFAULT true,
    sort_order INT     NOT NULL DEFAULT 0,
    UNIQUE (command_id, name)
);

CREATE TABLE command_permissions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    command_id   UUID NOT NULL REFERENCES commands(id) ON DELETE CASCADE,
    grantee_type TEXT NOT NULL CHECK (grantee_type IN ('user', 'group')),
    grantee_id   UUID NOT NULL,
    UNIQUE (command_id, grantee_type, grantee_id)
);

CREATE TABLE config_entries (
    id              UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    key             TEXT        NOT NULL UNIQUE,
    value_encrypted BYTEA       NOT NULL,
    description     TEXT        NOT NULL DEFAULT '',
    created_by      UUID        NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE executions (
    id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    command_id    UUID        NOT NULL REFERENCES commands(id),
    triggered_by  UUID        NOT NULL REFERENCES users(id),
    inputs        JSONB       NOT NULL DEFAULT '{}',
    status        TEXT        NOT NULL CHECK (status IN ('pending', 'running', 'success', 'failure')) DEFAULT 'pending',
    error_message TEXT        NOT NULL DEFAULT '',
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down

DROP TABLE executions;
DROP TABLE config_entries;
DROP TABLE command_permissions;
DROP TABLE command_inputs;
DROP TABLE commands;
DROP TABLE user_group_members;
DROP TABLE groups;
DROP TABLE users;
