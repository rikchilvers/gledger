package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	// flag to pass the journal file to read
	rootJournalPath string
	// flag to include only transactions on or after this date
	beginDate string
	// flag to include only transactions before this date
	endDate string
	// flag to include only transactions on or before today
	current       bool
	filterContext filteringContext
)

type filteringContext struct {
	checkPayees   bool
	checkAccounts bool
	checkNotes    bool
}

func newFilteringContext(checkPayees, checkAccounts, checkNotes bool) filteringContext {
	return filteringContext{
		checkPayees:   checkPayees,
		checkAccounts: checkAccounts,
		checkNotes:    checkNotes,
	}
}

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
	rootCmd.PersistentFlags().BoolVarP(&current, "current", "c", false, "include only transactions on or before today (overrides --begin and --end)")

	filterContext = newFilteringContext(true, true, true)
}

// Execute runs gledger
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
