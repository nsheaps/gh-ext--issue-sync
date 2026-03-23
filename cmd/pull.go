package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pullAll bool

var pullCmd = &cobra.Command{
	Use:   "pull [issue-number]",
	Short: "Pull issues from GitHub to local files",
	Long: `Pull one or more GitHub issues and save them as local markdown files
with YAML frontmatter containing issue metadata.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pullAll {
			return runPullAll()
		}
		if len(args) == 0 {
			return fmt.Errorf("specify an issue number or use --all")
		}
		return runPull(args[0])
	},
}

func init() {
	pullCmd.Flags().BoolVar(&pullAll, "all", false, "Pull all issues")
}

func runPull(issueNumber string) error {
	fmt.Printf("pull issue #%s: not yet implemented\n", issueNumber)
	return nil
}

func runPullAll() error {
	fmt.Println("pull all issues: not yet implemented")
	return nil
}
