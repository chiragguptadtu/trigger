# Project Trigger

Internal ops platform. Wraps Python/bash scripts into a web UI so non-technical staff can run them without engineering involvement.

## Roles
- **Operator** (default): run assigned commands, view full execution history for those commands
- **Admin**: everything + user/group/config/permission management, implicit access to all commands

## Command Registration
One file per command in `commands/`. Metadata lives in a `# ---trigger---` / `# ---end---` YAML comment header. Two input types: `open` (free text), `closed` (select from options, supports `multi: true`). Execution contract: empty string returned = success, non-empty = error message. Files without a trigger block are flagged as import errors.

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
| Backend | Go, `net/http` (stdlib, no framework), pgx/v5, sqlc, goose, river, golang-jwt, x/crypto |
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
  command/          # ParseContent(), ScanDir(), ScanLoop() — script discovery
  config/           # app config (env vars)
  crypto/           # AES-256-GCM encrypt/decrypt helpers
  execution/        # River jobs, worker, subprocess runner
  handler/          # HTTP handlers, Server struct, typed response structs
  middleware/       # JWT auth middleware, admin-only guard, command access guard
  store/            # sqlc-generated type-safe DB queries (do not edit manually)
commands/           # user-defined command scripts
web/
  src/
    api/            # Axios clients: auth, commands, executions, admin
    components/     # Navbar, Sidebar, CommandDetail, Brand
    layouts/        # AppLayout (sidebar + content shell)
    pages/          # LoginPage, SignupPage, CommandsPage, Admin* pages
    utils/          # table.tsx (tableProps/colTitle), theme.ts (PRIMARY/HOVER_BG/AVATAR_COLORS),
                    # error.ts (getApiError), jwt.ts (decodeClaims)
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

**Frontend conventions:**
- All tables use shared `tableProps` and `colTitle()` from `utils/table.tsx`
- Theme colors defined once in `utils/theme.ts` — import `PRIMARY`, `HOVER_BG`, `AVATAR_COLORS`
- API errors extracted via `getApiError(err, fallback)` from `utils/error.ts`
- Action buttons use Ant Design icons, never text labels
- Tooltips: `color="#fff"` with `styles={{ body: { color: 'rgba(0,0,0,0.65)' } }}`

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
- [x] Frontend (React + TypeScript) — all screens complete

## API Shape
```
POST   /auth/login

GET    /commands                        # operator: own; admin: all
GET    /commands/{slug}
POST   /commands/{slug}/executions      # trigger a command
GET    /commands/{slug}/executions      # history for current user
GET    /executions/{id}

# Admin only
CRUD   /admin/users
CRUD   /admin/groups
CRUD   /admin/groups/{id}/members
CRUD   /admin/commands/{slug}/permissions
GET    /admin/commands/import-errors
CRUD   /admin/config
```
