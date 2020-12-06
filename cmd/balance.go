package cmd

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

var (
	flattenTree bool
	showZero    bool
)

var balanceCmd = &cobra.Command{
	Use:          "balance",
	Aliases:      []string{"bal", "b"},
	Short:        "Shows accounts and their balances",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, _ []string) {
		journal := journal.NewJournal()
		if err := parse(journal.AddTransaction, nil); err != nil {
			fmt.Println(err)
			return
		}
		journal.Prepare(showZero)
		journal.Report(flattenTree)
	},
}

func init() {
	balanceCmd.Flags().BoolVarP(&flattenTree, "flat", "l", false, "show accounts as a flat list")
	balanceCmd.Flags().BoolVarP(&showZero, "empty", "E", false, "show accounts with zero amount")
	rootCmd.AddCommand(balanceCmd)
}
