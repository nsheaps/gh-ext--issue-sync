package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show sync status of local issue files",
	Long:  `Compare local issue files against GitHub and show which are modified, new, or deleted.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStatus()
	},
}

func runStatus() error {
	fmt.Println("status: not yet implemented")
	return nil
}
