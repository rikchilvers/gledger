package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of gledger",
	Long:  `All software has versions. This is gledger's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("gledger v0.1")
	},
}
