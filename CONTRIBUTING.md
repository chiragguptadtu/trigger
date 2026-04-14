# Contributing to Trigger

Thanks for your interest in contributing. Here's how to get started.

## Prerequisites

- Go 1.25+
- Node.js 20+
- Docker (for PostgreSQL)

## Setup

```bash
git clone https://github.com/chiragguptadtu/trigger.git
cd trigger
cp .env.example .env   # fill in at minimum JWT_SECRET and ENCRYPTION_KEY
docker compose up -d
go build ./...
```

Start the frontend dev server:

```bash
cd web && npm install && npm run dev
```

## Running Tests

```bash
# Backend
docker compose up -d   # PostgreSQL must be running
go test ./...

# Frontend
cd web && npm run test
```

Backend tests require a running PostgreSQL instance. Migrations are applied automatically by the test suite.

## Project Structure

```
cmd/server/         # main entry point
db/
  migrations/       # goose SQL migrations (source of truth)
  queries/          # sqlc query files
internal/
  auth/             # JWT and password helpers
  command/          # script scanner and sync loop
  config/           # env var loading
  crypto/           # AES-256-GCM helpers
  execution/        # River job worker and subprocess runner
  handler/          # HTTP handlers and typed response structs
  middleware/       # JWT auth and access control
  store/            # sqlc-generated DB layer (do not edit manually)
commands/           # user-defined command scripts
web/
  src/
    api/            # Axios API clients (auth, commands, executions, admin)
    components/     # Shared components (Navbar, Sidebar, CommandDetail, Brand)
    layouts/        # AppLayout (sidebar + content shell)
    pages/          # Route-level pages (Login, Commands, Admin*)
    utils/          # Shared utilities (table config, theme constants, error helper, JWT decode)
```

## Making Changes

### Adding a database migration

1. Add `NNN_description.sql` to `db/migrations/`
2. Regenerate the store: `cd db && sqlc generate`
3. Migrations are applied automatically on server start

### Regenerating the store after query changes

```bash
cd db && sqlc generate
```

### Code style

**Backend:**
- No frameworks — stdlib `net/http` only
- Typed response structs in `internal/handler/response.go`; never return raw DB types
- Partial PATCH updates using pointer fields
- Handler tests use `httptest` + real DB; store tests use rollback transactions

**Frontend:**
- All tables use shared `tableProps` and `colTitle()` from `web/src/utils/table.tsx`
- Theme colors (primary, hover, avatar) are defined once in `web/src/utils/theme.ts`
- API error messages are extracted via `getApiError()` from `web/src/utils/error.ts`
- Action buttons use Ant Design icons (`@ant-design/icons`), never text labels
- Tooltips use white background: `color="#fff"` with `styles={{ body: { color: 'rgba(0,0,0,0.65)' } }}`

## Submitting a Pull Request

1. Fork the repository
2. Create a branch: `git checkout -b your-feature`
3. Make your changes and ensure `go test ./...` passes
4. Open a pull request against `main` with a clear description of the change

## License

By contributing, you agree that your contributions will be licensed under the same [Elastic License 2.0](LICENSE) as the rest of the project.
