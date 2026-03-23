package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var pushAll bool

var pushCmd = &cobra.Command{
	Use:   "push [issue-number]",
	Short: "Push local issue files back to GitHub",
	Long: `Push local markdown issue files back to GitHub, updating the issue
title, body, labels, assignees, and other metadata from the frontmatter.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pushAll {
			return runPushAll()
		}
		if len(args) == 0 {
			return fmt.Errorf("specify an issue number or use --all")
		}
		return runPush(args[0])
	},
}

func init() {
	pushCmd.Flags().BoolVar(&pushAll, "all", false, "Push all modified issues")
}

func runPush(issueNumber string) error {
	fmt.Printf("push issue #%s: not yet implemented\n", issueNumber)
	return nil
}

func runPushAll() error {
	fmt.Println("push all issues: not yet implemented")
	return nil
}
