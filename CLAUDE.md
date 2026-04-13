# Project Trigger

Internal ops platform. Wraps Python/bash scripts into a web UI so non-technical staff can run them without engineering involvement.

## Roles
- **Operator** (default): run assigned commands, view own execution history
- **Admin**: everything + user/group/config/permission management, implicit access to all commands

## Command Registration
One file per command in `commands/`. Metadata lives in a `# ---trigger---` / `# ---end---` YAML comment header. Two input types: `open` (free text), `closed` (select from options, supports `multi: true`). Execution contract: empty string returned = success, non-empty = error message.

```python
# ---trigger---
# name: My Command
# description: What it does
# inputs:
#   - name: env
#     label: Environment
#     type: closed
#     options: [staging, production]
#     multi: false
#     required: true
#   - name: reason
#     label: Reason
#     type: open
#     required: true
# ---end---
```

## Access Control
Grant access per **user** or per **group** only — no role-based grants. Admins bypass all grants.

## Tech Stack
| Layer | Choice |
|---|---|
| Backend | Go, `net/http` (stdlib, no framework), pgx/v5, sqlc, goose, river, golang-jwt, x/crypto, validator |
| DB | PostgreSQL (Docker Compose) |
| Frontend | React + TypeScript, Ant Design, TanStack Query, Vitest + RTL |

## Folder Structure
```
cmd/server/         # main entry point
db/
  db.go             # Connect() + RunMigrations() — migrations embedded here
  migrations/       # goose SQL migration files (source of truth)
  queries/          # SQL query files for sqlc
  sqlc.yaml
internal/
  auth/             # HashPassword, ComparePassword, GenerateToken, ValidateToken
  command/          # ParseContent() — parses trigger header block from script files
  config/           # app config (env vars)
  execution/        # River jobs, worker, subprocess runner
  handler/          # HTTP handlers, Server struct
  middleware/        # JWT auth middleware, admin-only guard
  store/            # sqlc-generated type-safe DB queries (do not edit manually)
commands/           # user-defined command scripts (example included)
```

## Key Patterns

**Store tests** use rollback transactions for isolation:
```go
withTx(t, func(q *store.Queries) { ... }) // always rolls back
```

**Handler tests** use `httptest` + shared pool + `uniqueEmail()` to avoid constraint collisions.

**Adding a migration:**
1. Add `NNN_description.sql` to `db/migrations/`
2. Re-run `cd db && sqlc generate` to regenerate store
3. `RunMigrations()` applies it automatically on server start

**Regenerate store after query changes:**
```bash
cd db && PATH=$PATH:~/go/bin sqlc generate
```

**Run all tests:**
```bash
docker compose up -d   # PostgreSQL must be running
go test trigger/...
```

## Implementation Status
- [x] Project setup (go.mod, Docker Compose, folder structure)
- [x] DB schema + migrations (goose + River) + sqlc codegen
- [x] Auth: HashPassword, JWT generate/validate, POST /auth/login
- [x] Command scanner: ScanDir (resilient — collects per-file errors), Sync to DB
- [x] Command auto-discovery: ScanLoop runs every 30s, persists parse errors to command_import_errors
- [x] Auth middleware (Authenticate, RequireAdmin, RequireCommandAccess)
- [x] Config/secrets CRUD (AES-GCM encryption, admin-only)
- [x] Execution engine: RunScript (py/sh subprocess), River worker, enqueuer, POST+GET handlers
- [x] Admin HTTP endpoints (users, groups, permissions, import errors)
- [ ] Frontend (React + TypeScript)

## API Shape (planned)
```
POST   /auth/login
GET    /commands                        # operator: own; admin: all
GET    /commands/{slug}
POST   /executions                      # trigger a command
GET    /executions/{id}
GET    /commands/{slug}/executions      # history for current user
CRUD   /admin/users
CRUD   /admin/groups
CRUD   /admin/groups/{id}/members
CRUD   /admin/commands/{slug}/permissions
GET    /admin/commands/import-errors    # parse errors from auto-discovery scan
CRUD   /admin/config
```
