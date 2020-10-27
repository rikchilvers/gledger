package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:     "stats",
	Aliases: []string{"statistics", "s"},
	Short:   "Shows some journal statistics",
	Run: func(cmd *cobra.Command, args []string) {
		gatherStatistics()
	},
}

func gatherStatistics() {
	fmt.Println("some statistics")
}
