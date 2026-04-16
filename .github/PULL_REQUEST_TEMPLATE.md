## Type
- [ ] Bug fix
- [ ] Feature
- [ ] Refactor
- [ ] Migration / schema change
- [ ] Docs / config

## Description
<!-- What does this PR do and why? Link any related issue. -->

## Checklist

### General
- [ ] CI passes (`go build ./...`, `go test -p 1 ./...`, `npm run build`)
- [ ] No debug code, commented-out blocks, or untracked TODOs left behind

### Backend — skip if no Go changes
- [ ] Store tests use `withTx` for rollback isolation
- [ ] Handler tests use `httptest` + `uniqueEmail()` to avoid constraint collisions
- [ ] Handlers go through the store layer — no raw pgx in handler code
- [ ] All HTTP responses use typed response structs (no raw DB types returned)
- [ ] PATCH endpoints update only provided fields (pointer fields, not replace-all)

### Database — skip if no schema changes
- [ ] New migration file added to `db/migrations/` with correct sequence number
- [ ] `cd db && sqlc generate` re-run and generated files committed
- [ ] Migration is additive — no destructive changes to existing live columns

### Command Scripts — skip if no changes to `commands/`
- [ ] `# ---trigger---` / `# ---end---` YAML header is present and valid
- [ ] Script returns empty string on success, non-empty error message on failure

### Frontend — skip if no frontend changes
- [ ] `npm run build` passes (includes `tsc -b` type check)
- [ ] `npm run lint` passes with no warnings
- [ ] Tables use `tableProps` + `colTitle()` from `utils/table.tsx`
- [ ] Colors imported from `utils/theme.ts` — no hardcoded hex values
- [ ] API errors extracted via `getApiError()` from `utils/error.ts`
- [ ] Action buttons use Ant Design icons, not text labels
- [ ] Tooltips use `color="#fff"` with `styles={{ body: { color: 'rgba(0,0,0,0.65)' } }}`

## Screenshots
<!-- Required for any UI changes — include before and after -->

## Testing notes
<!-- Describe how you tested this change -->
