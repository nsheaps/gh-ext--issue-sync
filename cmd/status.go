package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nsheaps/gh-ext--issue-sync/internal/frontmatter"
	"github.com/nsheaps/gh-ext--issue-sync/internal/sync"
	"github.com/spf13/cobra"
)

var statusDir string

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status of local issue files",
	Long:  `Compare local issue files against GitHub and show which are modified, new, or deleted.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		repo, err := sync.ResolveRepo()
		if err != nil {
			return err
		}
		return runStatus(repo)
	},
}

func init() {
	statusCmd.Flags().StringVarP(&statusDir, "dir", "d", "issues", "Directory containing issue files")
}

func runStatus(repo string) error {
	// Read local files
	pattern := filepath.Join(statusDir, "*.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return fmt.Errorf("listing issue files: %w", err)
	}

	if len(files) == 0 {
		fmt.Println("No local issue files found. Run 'gh ext-issue-sync pull --all' to get started.")
		return nil
	}

	fmt.Fprintf(os.Stderr, "Comparing %d local files against %s...\n\n", len(files), repo)

	for _, filename := range files {
		f, err := os.Open(filename)
		if err != nil {
			return fmt.Errorf("opening %s: %w", filename, err)
		}

		var meta sync.IssueFrontmatter
		localBody, err := frontmatter.Unmarshal(f, &meta)
		f.Close()
		if err != nil {
			return fmt.Errorf("parsing %s: %w", filename, err)
		}

		localBody = strings.TrimSpace(localBody)

		// Fetch remote issue
		remote, err := sync.FetchIssue(repo, meta.Number)
		if err != nil {
			fmt.Printf("  ? #%d %s (could not fetch from GitHub)\n", meta.Number, meta.Title)
			continue
		}

		// Compare key fields
		changes := []string{}
		if meta.Title != remote.Title {
			changes = append(changes, "title")
		}
		if localBody != strings.TrimSpace(remote.Body) {
			changes = append(changes, "body")
		}
		if meta.State != remote.State {
			changes = append(changes, "state")
		}
		if !stringSliceEqual(meta.Labels, remote.Labels) {
			changes = append(changes, "labels")
		}
		if !stringSliceEqual(meta.Assignees, remote.Assignees) {
			changes = append(changes, "assignees")
		}

		if len(changes) == 0 {
			fmt.Printf("  = #%d %s\n", meta.Number, meta.Title)
		} else {
			fmt.Printf("  M #%d %s [%s]\n", meta.Number, meta.Title, strings.Join(changes, ", "))
		}
	}

	return nil
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
