# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-04-14

### Added
- Zero-touch command registration via `# ---trigger---` YAML comment headers
- Two roles: Operator (run assigned commands, view execution history) and Admin (full access)
- Fine-grained access control — grant per user or per group; admins bypass all grants
- Async execution via River job queue with live status polling
- AES-256-GCM encrypted config/secrets store injected into scripts at runtime
- Import error surfacing — broken command files flagged in UI without affecting healthy commands
- Admin password reset — reset any user's password; self-lockout recovery via startup env vars
- React + TypeScript frontend with Ant Design, TanStack Query
- JWT authentication with bcrypt password hashing
- Full admin CRUD: users, groups, group membership, command permissions, config entries
- Execution history per command with stdout/stderr capture
