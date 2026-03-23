# gh-ext--issue-sync

GitHub CLI extension that syncs issues to local markdown files with YAML
frontmatter.

## Project Structure

- `main.go` — Entry point
- `cmd/` — Cobra command definitions (root, pull, push, status)
- `cmd/cmd_test.go` — Command tests using mock client
- `internal/sync/` — Core sync logic, Client interface, and GHClient (gh CLI)
- `internal/sync/client.go` — Client interface (abstraction for testability)
- `internal/sync/github.go` — GHClient: real implementation using `gh` CLI
- `internal/sync/push.go` — Push logic with milestone title-to-number resolution
- `internal/sync/issue.go` — Issue and IssueFrontmatter types
- `internal/frontmatter/` — Generic YAML frontmatter marshal/unmarshal
- `mise/tasks/` — Build, test, lint tasks (bash scripts with `#MISE` pragmas)
- `.github/workflows/` — CI (build+test+lint) and release (precompile on tag)

## Development

- **Build**: `mise run build` or `go build ./...`
- **Test**: `mise run test` or `go test -cover ./...`
- **Lint**: `mise run lint` (go vet + prettier)
- **All checks**: `mise run check`

## Conventions

- Go code uses standard `go vet` linting
- Non-Go files use prettier for formatting
- Mise tasks are the source of truth for all build/lint/test logic
- CI workflows are thin wrappers around mise tasks
- Commit messages follow conventional commits (`feat:`, `fix:`, `chore:`,
  `docs:`)

## Architecture Decisions

- **Client interface** (`sync.Client`): All GitHub operations go through this
  interface. The real `GHClient` shells out to `gh` CLI. Tests use a mock
  client. This keeps commands testable without network access.
- **Frontmatter is generic**: The `frontmatter` package knows nothing about
  GitHub. It just serializes/deserializes YAML frontmatter + markdown body.
- **Pull requests are filtered**: GitHub's issues API returns PRs too. We filter
  them out via the `pull_request` field.
- **Milestone resolution**: Push resolves milestone titles to numbers via the
  milestones API. Status compares milestone titles directly.
- **Label/assignee comparison is order-independent**: Status sorts before
  comparing to avoid false positives from API ordering.
