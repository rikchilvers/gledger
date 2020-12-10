package cmd

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

var budgetCmd = &cobra.Command{
	Use:          "budget",
	Aliases:      []string{"bud", "B"},
	Short:        "Shows budget accounts and their balances",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, _ []string) {
		config := journal.JournalConfig{
			CalculateBudget: true,
		}
		journal := journal.NewJournal(config)
		if err := parse(journal.AddTransaction, journal.AddPeriodicTransaction); err != nil {
			fmt.Println(err)
			return
		}
		journal.Prepare(showZero)
		report(journal.Root, flattenTree)
	},
}

func init() {
	budgetCmd.Flags().BoolVarP(&flattenTree, "flat", "l", false, "show accounts as a flat list")
	budgetCmd.Flags().BoolVarP(&showZero, "empty", "E", false, "show accounts with zero amount")
	rootCmd.AddCommand(budgetCmd)
}
