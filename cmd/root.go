package cmd

import (
	"github.com/spf13/cobra"
)

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
