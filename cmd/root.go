package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootJournalPath string

var rootCmd = &cobra.Command{
	Use:   "gledger",
	Short: "gledger - command line budgeting",
	Long:  "gledger is a reimplementation of Ledger in Go\nwith YNAB-style budgeting at its core",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootJournalPath, "file", "f", "", "journal file to read (defaults to $LEDGER_FILE)")
}

// Execute runs gledger
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
