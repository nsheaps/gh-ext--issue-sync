# gh-ext-issue-sync

A GitHub CLI extension that syncs GitHub issues to local markdown files with
YAML frontmatter.

Edit issues in your favorite editor, track them in git, and push changes back to
GitHub.

## Why?

- **Offline editing** — Work on issues without a browser. Edit titles, bodies,
  labels, and assignees in markdown.
- **Version control** — Track issue changes alongside code. See who changed what
  and when.
- **Bulk editing** — Update dozens of issues by editing files, then push them
  all at once.
- **AI-friendly** — Feed issue files to AI tools that work better with local
  files than APIs.
- **Backup** — Keep a local copy of all your issues as plain text.

## Installation

```bash
gh extension install nsheaps/gh-ext--issue-sync
```

Requires the [GitHub CLI](https://cli.github.com/) (`gh`) to be installed and
authenticated.

## Quick Start

```bash
# Pull all open issues from the current repo
gh ext-issue-sync pull --all

# Edit an issue in your editor
$EDITOR issues/00042.md

# See what changed
gh ext-issue-sync status

# Push changes back to GitHub
gh ext-issue-sync push 42
```

## Commands

### `pull` — Download issues from GitHub

```bash
# Pull a single issue
gh ext-issue-sync pull 42

# Pull all open issues
gh ext-issue-sync pull --all

# Pull closed issues
gh ext-issue-sync pull --all --state=closed

# Pull all issues (open + closed)
gh ext-issue-sync pull --all --state=all

# Pull into a custom directory
gh ext-issue-sync pull --all --dir=my-issues
```

### `push` — Upload local changes to GitHub

```bash
# Push a single issue
gh ext-issue-sync push 42

# Push all issues
gh ext-issue-sync push --all

# Push from a custom directory
gh ext-issue-sync push --all --dir=my-issues
```

When pushing, the following fields are updated on GitHub:

- Title
- Body (markdown content)
- State (open/closed)
- Labels
- Assignees
- Milestone (resolved by title)

### `status` — Compare local files against GitHub

```bash
gh ext-issue-sync status
```

Output shows each issue's sync state:

```
Comparing 5 local files against owner/repo...

  = #1 Setup CI pipeline
  M #2 Fix login bug [title, labels]
  = #3 Add dark mode
  M #4 Update docs [body]
  ? #5 Deleted issue (could not fetch from GitHub)
```

- `=` — In sync with GitHub
- `M` — Modified locally (shows which fields changed)
- `?` — Could not fetch from GitHub (deleted or permissions issue)

### Flags

All commands support:

| Flag        | Short | Default  | Description                           |
| ----------- | ----- | -------- | ------------------------------------- |
| `--dir`     | `-d`  | `issues` | Directory for issue files             |
| `--all`     |       |          | Process all issues                    |
| `--state`   |       | `open`   | Issue state filter (pull only)        |
| `--dry-run` |       |          | Show what would be pushed (push only) |
| `--help`    | `-h`  |          | Show help                             |

## File Format

Issues are stored as markdown files with YAML frontmatter. The filename is the
zero-padded issue number (e.g., `00042.md`).

```markdown
---
number: 42
title: Fix login bug
state: open
labels:
  - bug
  - priority/high
assignees:
  - octocat
milestone: v1.0
created_at: 2026-01-15T10:30:00Z
updated_at: 2026-03-20T14:22:00Z
author: octocat
---

The login form crashes when the password field is empty.

## Steps to Reproduce

1. Go to /login
2. Leave password blank
3. Click "Sign In"

## Expected Behavior

Show a validation error message.
```

### Frontmatter Fields

| Field        | Type     | Editable | Description                        |
| ------------ | -------- | -------- | ---------------------------------- |
| `number`     | int      | No       | GitHub issue number                |
| `title`      | string   | Yes      | Issue title                        |
| `state`      | string   | Yes      | `open` or `closed`                 |
| `labels`     | string[] | Yes      | Label names                        |
| `assignees`  | string[] | Yes      | GitHub usernames                   |
| `milestone`  | string   | Yes      | Milestone title (resolved on push) |
| `created_at` | datetime | No       | When the issue was created         |
| `updated_at` | datetime | No       | When the issue was last updated    |
| `author`     | string   | No       | Original author's username         |

Fields marked "No" in Editable are ignored during push — they reflect the state
at pull time.

## Examples

### Bulk-close issues

```bash
# Pull all open issues
gh ext-issue-sync pull --all

# Change state to closed in each file
for f in issues/*.md; do
  sed -i 's/^state: open/state: closed/' "$f"
done

# Push all changes
gh ext-issue-sync push --all
```

### Add a label to all issues

```bash
gh ext-issue-sync pull --all

# Add "needs-triage" label to each file (after the labels: line)
for f in issues/*.md; do
  sed -i '/^labels:/a\  - needs-triage' "$f"
done

gh ext-issue-sync push --all
```

### Track issues in git

```bash
# Pull and commit
gh ext-issue-sync pull --all
git add issues/
git commit -m "sync: pull latest issues"

# Make edits, then push to GitHub and commit
gh ext-issue-sync push --all
gh ext-issue-sync pull --all  # refresh timestamps
git add issues/
git commit -m "sync: update issue labels"
```

## Development

### Prerequisites

- [Go](https://go.dev/) 1.24+
- [mise](https://mise.jdx.dev/) (task runner and tool manager)
- [GitHub CLI](https://cli.github.com/) (`gh`)

### Setup

```bash
git clone https://github.com/nsheaps/gh-ext--issue-sync.git
cd gh-ext--issue-sync
mise install -y
```

### Tasks

```bash
mise run build      # Build the binary
mise run test       # Run tests with coverage
mise run lint       # Run linters (go vet + prettier)
mise run lint-check # Lint without auto-fix (CI mode)
mise run check      # Run all checks (lint + test + build)
mise run setup      # Install tools and download deps
```

### Project Structure

```
.
├── main.go                     # Entry point
├── cmd/
│   ├── root.go                 # Root command and client setup
│   ├── pull.go                 # Pull command
│   ├── push.go                 # Push command
│   ├── status.go               # Status command
│   └── cmd_test.go             # Command tests (mock client)
├── internal/
│   ├── frontmatter/
│   │   ├── frontmatter.go      # YAML frontmatter marshal/unmarshal
│   │   └── frontmatter_test.go # Frontmatter tests
│   └── sync/
│       ├── client.go           # Client interface
│       ├── github.go           # GHClient (gh CLI implementation)
│       ├── github_test.go      # API conversion tests
│       ├── issue.go            # Issue and IssueFrontmatter types
│       └── push.go             # Push logic and milestone resolution
├── mise/tasks/                 # Build, test, lint task scripts
├── .github/workflows/
│   ├── ci.yaml                 # CI: build, test, lint
│   └── release.yaml            # Release: precompile on tag
├── mise.toml                   # Tool versions
├── go.mod / go.sum             # Go module
└── CLAUDE.md                   # AI assistant context
```

### Architecture

The codebase follows a clean separation:

- **`internal/sync.Client`** — Interface for GitHub operations (fetch, push,
  resolve repo). The `GHClient` implementation shells out to `gh` CLI. Tests use
  a mock client.
- **`internal/frontmatter`** — Generic YAML frontmatter serialization,
  independent of GitHub.
- **`cmd/`** — Cobra command handlers that wire the client to file I/O.

### Running Tests

```bash
# All tests
go test ./...

# With verbose output
go test -v ./...

# With coverage
go test -cover ./...

# Coverage report
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Release

Releases are automated via GitHub Actions. To create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The release workflow builds precompiled binaries for all platforms using
[`cli/gh-extension-precompile`](https://github.com/cli/gh-extension-precompile).

## Troubleshooting

### "resolving repo: is gh authenticated and in a git repo?"

Make sure you're in a git repository with a GitHub remote, and that `gh` is
authenticated:

```bash
gh auth status
gh repo view
```

### "milestone not found"

The push command resolves milestone titles to numbers. If you see this error,
the milestone title in your frontmatter doesn't match any milestone on GitHub.
Check available milestones with:

```bash
gh api repos/OWNER/REPO/milestones --jq '.[].title'
```

### Push fails for some issues in --all mode

Push continues on errors and reports a summary at the end. Fix the failing
issues and re-run. The `--dry-run` flag lets you validate files before pushing:

```bash
gh ext-issue-sync push --all --dry-run
```

## Known Limitations

- **Pull requests are excluded** — GitHub's issues API returns PRs as issues.
  This tool filters them out automatically.
- **Read-only fields** — `number`, `created_at`, `updated_at`, and `author` are
  ignored during push. They reflect the state at pull time.
- **No comments sync** — Only the issue body and metadata are synced.
- **No conflict detection** — If an issue is modified on GitHub after you pull
  it, pushing will overwrite those changes. Use `status` to check before
  pushing.
- **Milestone resolution** — Milestones are matched by exact title. If you
  rename a milestone on GitHub, update the local files to match.

## License

MIT
