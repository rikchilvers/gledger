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
		config := journal.ProcessingConfig{
			CalculateBudget: true,
		}
		journal := journal.NewJournal(config)
		if err := parse(journal.AddTransaction, journal.AddPeriodicTransaction); err != nil {
			fmt.Println(err)
			return
		}
		journal.Prepare(showZero)
		report(*journal.BudgetRoot, flattenTree, collapseOnlyChildren)
	},
}

func init() {
	budgetCmd.Flags().BoolVarP(&flattenTree, "flatten", "F", false, "show accounts as a flat list")
	budgetCmd.Flags().BoolVarP(&showZero, "show-zero", "Z", false, "show accounts with zero amount")
	budgetCmd.Flags().BoolVarP(&collapseOnlyChildren, "collapse", "C", false, "collapse single child accounts into a list")
	rootCmd.AddCommand(budgetCmd)
}
