# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.0] - 2026-04-15

### Added
- Bash script support — `.sh` commands work end to end alongside Python (same input/config contract)
- Dynamic closed inputs (`dynamic: true`) — options resolved at scan time by calling `get_options(input_name, config)` in the script, refreshed every 30 s without a server restart
- Config store values decrypted and passed to `get_options` so scripts can drive options from live data
- `GetOptions` runner function with full test coverage for Python and bash
- Example commands: `ping_host.sh` (bash), `provision_environment.py` and `rotate_logs.sh` (dynamic inputs)

## [0.3.0] - 2026-04-15

### Changed
- Upgraded React 18 → 19, antd 5 → 6, vite 5 → 8 (Rolldown bundler), @ant-design/icons 5 → 6, @vitejs/plugin-react 4 → 6
- Node.js requirement bumped to 24 (vite 8 requires Node >= 20.19)
- Bundle size reduced from 1,123 kB to 1,075 kB due to improved tree-shaking in Rolldown

## [0.2.0] - 2026-04-14

### Changed
- Command form redesigned to a Google Sheets-like layout: labels as a fixed header row, inputs as a horizontally scrollable row — handles 15+ inputs cleanly without pushing execution history off screen
- Run button pinned below the scrollable input row, always visible
- Large hardcoded pixel values converted to viewport-relative units (`vw`) across all pages: sidebar width, form input width, modals, login card, permission selector, navbar popover
- Admin page padding centralized into a shared `PAGE_PADDING` constant

### Fixed
- Multi-select dropdowns in command forms no longer flip upward when many options are selected

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
