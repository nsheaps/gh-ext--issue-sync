package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/nsheaps/gh-ext--issue-sync/internal/frontmatter"
	"github.com/nsheaps/gh-ext--issue-sync/internal/sync"
	"github.com/spf13/cobra"
)

var (
	pullAll   bool
	pullState string
	pullDir   string
)

var pullCmd = &cobra.Command{
	Use:   "pull [issue-number]",
	Short: "Pull issues from GitHub to local files",
	Long: `Pull one or more GitHub issues and save them as local markdown files
with YAML frontmatter containing issue metadata.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := sync.ResolveRepo()
		if err != nil {
			return err
		}

		if pullAll {
			return runPullAll(repo)
		}
		if len(args) == 0 {
			return fmt.Errorf("specify an issue number or use --all")
		}
		num, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid issue number %q: %w", args[0], err)
		}
		return runPull(repo, num)
	},
}

func init() {
	pullCmd.Flags().BoolVar(&pullAll, "all", false, "Pull all issues")
	pullCmd.Flags().StringVar(&pullState, "state", "open", "Issue state filter: open, closed, all")
	pullCmd.Flags().StringVarP(&pullDir, "dir", "d", "issues", "Directory to store issue files")
}

func runPull(repo string, number int) error {
	issue, err := sync.FetchIssue(repo, number)
	if err != nil {
		return err
	}

	return writeIssueFile(issue)
}

func runPullAll(repo string) error {
	issues, err := sync.FetchAllIssues(repo, pullState)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Pulling %d issues from %s...\n", len(issues), repo)

	for _, issue := range issues {
		if err := writeIssueFile(issue); err != nil {
			return err
		}
	}

	fmt.Fprintf(os.Stderr, "Done. %d issues written to %s/\n", len(issues), pullDir)
	return nil
}

func writeIssueFile(issue *sync.Issue) error {
	if err := os.MkdirAll(pullDir, 0o755); err != nil {
		return fmt.Errorf("creating directory %s: %w", pullDir, err)
	}

	filename := fmt.Sprintf("%s/%05d.md", pullDir, issue.Number)

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", filename, err)
	}
	defer f.Close()

	meta := sync.IssueFrontmatter{
		Number:    issue.Number,
		Title:     issue.Title,
		State:     issue.State,
		Labels:    issue.Labels,
		Assignees: issue.Assignees,
		Milestone: issue.Milestone,
		CreatedAt: issue.CreatedAt,
		UpdatedAt: issue.UpdatedAt,
		Author:    issue.Author,
	}

	if err := frontmatter.Marshal(f, &meta, issue.Body); err != nil {
		return fmt.Errorf("writing issue #%d: %w", issue.Number, err)
	}

	fmt.Fprintf(os.Stderr, "  %s (#%d: %s)\n", filename, issue.Number, issue.Title)
	return nil
}
