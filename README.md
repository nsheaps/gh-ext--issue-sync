# gh-ext--issue-sync

A GitHub CLI extension that syncs GitHub issues to local markdown files with frontmatter metadata.

## Installation

```bash
gh extension install nsheaps/gh-ext--issue-sync
```

## Usage

```bash
# Pull a single issue to a local file
gh ext-issue-sync pull 42

# Pull all issues
gh ext-issue-sync pull --all

# Push local changes back to GitHub
gh ext-issue-sync push 42

# Push all modified issues
gh ext-issue-sync push --all

# Check sync status
gh ext-issue-sync status
```

## File Format

Issues are stored as markdown files with YAML frontmatter:

```markdown
---
number: 42
title: "Fix login bug"
state: open
labels:
  - bug
  - priority/high
assignees:
  - username
milestone: "v1.0"
created_at: "2026-01-15T10:30:00Z"
updated_at: "2026-03-20T14:22:00Z"
author: "octocat"
---

Issue body content here as markdown...
```

## Development

```bash
# Install tools
mise install -y

# Build
mise run build

# Run tests
mise run test

# Lint
mise run lint
```
