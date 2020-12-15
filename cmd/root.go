package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	rootJournalPath string
	beginDate       string
	endDate         string
)

var rootCmd = &cobra.Command{
	Use:   "gledger",
	Short: "gledger - command line budgeting",
	Long:  "gledger is a reimplementation of Ledger in Go\nwith YNAB-style budgeting at its core",
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootJournalPath, "file", "f", "", "journal file to read (default $LEDGER_FILE)")
	rootCmd.PersistentFlags().StringVarP(&beginDate, "begin", "b", "", "include only transactions on or after this date")
	rootCmd.PersistentFlags().StringVarP(&endDate, "end", "e", "", "include only transactions before this date")
}

// Execute runs gledger
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
