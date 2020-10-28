package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:          "stats",
	Aliases:      []string{"statistics", "s"},
	Short:        "Shows some journal statistics",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return gatherStatistics()
	},
}

func gatherStatistics() error {
	return parse()
}
