package cmd

import (
	"github.com/nsheaps/gh-ext--issue-sync/internal/sync"
	"github.com/spf13/cobra"
)

// ghClient is the GitHub client used by all commands.
// Defaults to the real gh CLI client; overridden in tests.
var ghClient sync.Client = sync.NewGHClient()

var rootCmd = &cobra.Command{
	Use:   "gh-ext-issue-sync",
	Short: "Sync GitHub issues to local markdown files",
	Long: `gh-ext-issue-sync synchronizes GitHub issues to local markdown files
with YAML frontmatter. Supports push/pull for individual issues or entire repos.`,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(statusCmd)
}
