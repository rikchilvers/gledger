package cmd

import (
	"os"

	"github.com/rikchilvers/gledger/reporting"
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
	current bool
	filters []reporting.Filter
)

var rootCmd = &cobra.Command{
	Use:   "gledger",
	Short: "gledger - command line budgeting",
	Long:  "gledger is a reimplementation of Ledger\nwith YNAB-style budgeting at its core",
	PersistentPreRunE: func(_ *cobra.Command, args []string) error {
		filters = make([]reporting.Filter, 0, len(args))
		for _, arg := range args {
			filter, err := reporting.NewFilter(arg)
			if err != nil {
				return err
			}
			filters = append(filters, filter)
		}
		return nil
	},
	Run: func(cmd *cobra.Command, _ []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootJournalPath, "file", "f", "", "journal file to read (default $LEDGER_FILE)")
	rootCmd.PersistentFlags().StringVarP(&beginDate, "begin", "b", "", "include only transactions on or after this date")
	rootCmd.PersistentFlags().StringVarP(&endDate, "end", "e", "", "include only transactions before this date")
	rootCmd.PersistentFlags().BoolVarP(&current, "current", "c", false, "include only transactions on or before today (overrides --begin and --end)")
}

// Execute runs gledger
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
