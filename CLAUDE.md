# gh-ext--issue-sync

GitHub CLI extension that syncs issues to local markdown files with YAML frontmatter.

## Project Structure

- `main.go` — Entry point
- `cmd/` — Cobra command definitions (root, pull, push, status)
- `internal/sync/` — Core sync logic and issue types
- `internal/frontmatter/` — YAML frontmatter marshal/unmarshal
- `mise/tasks/` — Build, test, lint tasks (bash scripts with `#MISE` pragmas)
- `.github/workflows/` — CI and release workflows

## Development

- **Build**: `mise run build` or `go build ./...`
- **Test**: `mise run test` or `go test ./...`
- **Lint**: `mise run lint` (go vet + prettier)
- **All checks**: `mise run check`

## Conventions

- Go code uses standard `go vet` linting
- Non-Go files use prettier for formatting
- Mise tasks are the source of truth for all build/lint/test logic
- CI workflows are thin wrappers around mise tasks
- Commit messages follow conventional commits (`feat:`, `fix:`, `chore:`, `docs:`)
