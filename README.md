# Trigger

An open-source internal ops platform that wraps Python and Bash scripts into a secure web UI — so non-technical teammates can run them without engineering involvement.

> Think of it as a self-hosted alternative to tools like Airplane or Retool Workflows, minus the vendor lock-in and the managed-service pricing.

## Why Trigger?

Most teams accumulate scripts — deployment helpers, data fixes, report generators — that only engineers can run because they live on the command line. Trigger solves this by giving each script a form-based UI, access controls, and an execution history, all without touching the script itself.

## Features

- **Zero-touch command registration** — drop a script with a YAML comment header into the `commands/` directory; Trigger picks it up automatically within 30 seconds
- **Two roles** — Operator (run assigned commands, view own history) and Admin (manage everything)
- **Fine-grained access control** — grant access per user or per group; admins always have full access
- **Async execution** — commands run in the background via a River job queue; poll for results
- **Encrypted config store** — AES-256-GCM encrypted key-value pairs injected into scripts at runtime
- **Import error surfacing** — broken command files are flagged in the UI without affecting healthy commands
- **Admin password reset** — admins can reset any user's password; self-lockout recovery via startup env vars

## Tech Stack

| Layer | Choice |
|---|---|
| Backend | Go, `net/http` (stdlib), pgx/v5, sqlc, goose, River |
| Auth | JWT (golang-jwt), bcrypt (x/crypto) |
| Database | PostgreSQL |
| Frontend | React + TypeScript *(in progress)* |

## Getting Started

### Prerequisites

- Go 1.25+
- Docker (for PostgreSQL)

### 1. Clone and configure

```bash
git clone https://github.com/chiragguptadtu/trigger.git
cd trigger
cp .env.example .env
# Edit .env — at minimum set JWT_SECRET and ENCRYPTION_KEY
```

Generate a random encryption key:

```bash
openssl rand -hex 32
```

### 2. Start the database

```bash
docker compose up -d
```

### 3. Run the server

```bash
go run ./cmd/server
```

Migrations run automatically on startup. On first run, set `ADMIN_EMAIL` and `ADMIN_PASSWORD` in `.env` to seed an admin account.

The server starts on `http://localhost:8080`.

## Adding a Command

Create a Python or Bash script in the `commands/` directory with a `# ---trigger---` YAML header:

```python
# ---trigger---
# name: Deploy Service
# description: Deploys a service to an environment
# inputs:
#   - name: environment
#     label: Environment
#     type: closed
#     options: [staging, production]
#     required: true
#   - name: reason
#     label: Reason for deploy
#     type: open
#     required: true
# ---end---

import sys, json

def run(inputs, config):
    env = inputs["environment"]
    print(f"Deploying to {env}...")
    # your logic here
    return ""  # empty string = success; non-empty = error message

if __name__ == "__main__":
    inputs = json.loads(sys.argv[1]) if len(sys.argv) > 1 else {}
    config = json.loads(sys.argv[2]) if len(sys.argv) > 2 else {}
    result = run(inputs, config)
    if result:
        print(result, file=sys.stderr)
        sys.exit(1)
```

Trigger auto-discovers the script within 30 seconds. No restart required.

## Configuration

| Variable | Required | Description |
|---|---|---|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `JWT_SECRET` | Yes | Secret for signing JWTs |
| `ENCRYPTION_KEY` | Yes | 64-char hex string (32-byte AES-256 key) |
| `ADMIN_EMAIL` | First run | Seeds an admin user if the DB is empty |
| `ADMIN_PASSWORD` | First run | Password for the seeded admin |
| `RESET_ADMIN_EMAIL` | Recovery | Resets this user's password on startup |
| `RESET_ADMIN_PASSWORD` | Recovery | New password to set on startup |
| `COMMANDS_DIR` | No | Path to commands directory (default: `./commands`) |
| `PORT` | No | HTTP port (default: `8080`) |

## API Overview

```
POST   /auth/login

GET    /commands                          # list accessible commands
GET    /commands/{slug}
POST   /commands/{slug}/executions        # trigger a command
GET    /commands/{slug}/executions        # execution history (current user)
GET    /executions/{id}

# Admin only
CRUD   /admin/users
CRUD   /admin/groups
CRUD   /admin/groups/{id}/members
CRUD   /admin/commands/{slug}/permissions
GET    /admin/commands/import-errors
CRUD   /admin/config
```

## Running Tests

```bash
docker compose up -d   # PostgreSQL must be running
go test ./...
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

Elastic License 2.0 — free to use, modify, and contribute; you may not offer Trigger as a managed/hosted service. See [LICENSE](LICENSE) for full terms.
