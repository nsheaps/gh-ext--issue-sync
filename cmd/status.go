package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
		repo, err := ghClient.ResolveRepo()
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
		remote, err := ghClient.FetchIssue(repo, meta.Number)
		if err != nil {
			fmt.Printf("  ? #%d %s (could not fetch from GitHub)\n", meta.Number, meta.Title)
			continue
		}

		// Compare key fields
		var changes []string
		if meta.Title != remote.Title {
			changes = append(changes, "title")
		}
		if localBody != strings.TrimSpace(remote.Body) {
			changes = append(changes, "body")
		}
		if meta.State != remote.State {
			changes = append(changes, "state")
		}
		if !stringSliceEqualUnordered(meta.Labels, remote.Labels) {
			changes = append(changes, "labels")
		}
		if !stringSliceEqualUnordered(meta.Assignees, remote.Assignees) {
			changes = append(changes, "assignees")
		}
		if meta.Milestone != remote.Milestone {
			changes = append(changes, "milestone")
		}

		if len(changes) == 0 {
			fmt.Printf("  = #%d %s\n", meta.Number, meta.Title)
		} else {
			fmt.Printf("  M #%d %s [%s]\n", meta.Number, meta.Title, strings.Join(changes, ", "))
		}
	}

	return nil
}

// stringSliceEqualUnordered compares two string slices ignoring order.
func stringSliceEqualUnordered(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	aSorted := make([]string, len(a))
	bSorted := make([]string, len(b))
	copy(aSorted, a)
	copy(bSorted, b)
	sort.Strings(aSorted)
	sort.Strings(bSorted)
	for i := range aSorted {
		if aSorted[i] != bSorted[i] {
			return false
		}
	}
	return true
}
