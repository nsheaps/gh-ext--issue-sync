package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nsheaps/gh-ext--issue-sync/internal/frontmatter"
	"github.com/nsheaps/gh-ext--issue-sync/internal/sync"
	"github.com/spf13/cobra"
)

var (
	pushAll bool
	pushDir string
)

var pushCmd = &cobra.Command{
	Use:   "push [issue-number]",
	Short: "Push local issue files back to GitHub",
	Long: `Push local markdown issue files back to GitHub, updating the issue
title, body, labels, assignees, and other metadata from the frontmatter.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := ghClient.ResolveRepo()
		if err != nil {
			return err
		}

		if pushAll {
			return runPushAll(repo)
		}
		if len(args) == 0 {
			return fmt.Errorf("specify an issue number or use --all")
		}
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue number %q: %w", args[0], err)
		}
		if num <= 0 {
			return fmt.Errorf("issue number must be positive, got %d", num)
		}
		return runPush(repo, num)
	},
}

func init() {
	pushCmd.Flags().BoolVar(&pushAll, "all", false, "Push all modified issues")
	pushCmd.Flags().StringVarP(&pushDir, "dir", "d", "issues", "Directory containing issue files")
}

func runPush(repo string, number int) error {
	filename := filepath.Join(pushDir, fmt.Sprintf("%05d.md", number))
	issue, err := readIssueFile(filename)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Pushing issue #%d (%s)...\n", issue.Number, issue.Title)
	if err := ghClient.PushIssue(repo, issue); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "  Updated issue #%d\n", issue.Number)
	return nil
}

func runPushAll(repo string) error {
	pattern := filepath.Join(pushDir, "*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("listing issue files: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no issue files found in %s/", pushDir)
	}

	fmt.Fprintf(os.Stderr, "Pushing %d issues to %s...\n", len(files), repo)

	pushed := 0
	var errors []string
	for _, filename := range files {
		issue, err := readIssueFile(filename)
		if err != nil {
			errors = append(errors, fmt.Sprintf("  %s: %v", filename, err))
			continue
		}

		if err := ghClient.PushIssue(repo, issue); err != nil {
			errors = append(errors, fmt.Sprintf("  #%d: %v", issue.Number, err))
			continue
		}

		fmt.Fprintf(os.Stderr, "  Updated issue #%d (%s)\n", issue.Number, issue.Title)
		pushed++
	}

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d issues pushed, %d errors:\n%s\n", pushed, len(errors), strings.Join(errors, "\n"))
		return fmt.Errorf("%d of %d issues failed to push", len(errors), len(files))
	}

	fmt.Fprintf(os.Stderr, "Done. %d issues pushed.\n", pushed)
	return nil
}

func readIssueFile(filename string) (*sync.Issue, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", filename, err)
	}
	defer f.Close()

	var meta sync.IssueFrontmatter
	body, err := frontmatter.Unmarshal(f, &meta)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", filename, err)
	}

	body = strings.TrimSpace(body)

	if meta.Number <= 0 {
		return nil, fmt.Errorf("%s: invalid or missing issue number in frontmatter", filename)
	}
	if meta.Title == "" {
		return nil, fmt.Errorf("%s: missing title in frontmatter", filename)
	}

	return &sync.Issue{
		Number:    meta.Number,
		Title:     meta.Title,
		State:     meta.State,
		Labels:    meta.Labels,
		Assignees: meta.Assignees,
		Milestone: meta.Milestone,
		CreatedAt: meta.CreatedAt,
		UpdatedAt: meta.UpdatedAt,
		Author:    meta.Author,
		Body:      body,
	}, nil
}
