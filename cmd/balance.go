package cmd

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

var (
	flattenTree bool
	showZero    bool
	showBudget  bool
)

var balanceCmd = &cobra.Command{
	Use:          "balance",
	Aliases:      []string{"bal", "b"},
	Short:        "Shows accounts and their balances",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, _ []string) {
		config := journal.Config{
			CalculateBudget: showBudget,
		}
		journal := journal.NewJournal(config)
		if err := parse(journal.AddTransaction, journal.AddPeriodicTransaction); err != nil {
			fmt.Println(err)
			return
		}
		journal.Prepare(showZero)
		report(journal.Root, flattenTree)

		if showBudget {
			fmt.Println("")
			report(journal.BudgetRoot, flattenTree)
		}
	},
}

func init() {
	balanceCmd.Flags().BoolVarP(&flattenTree, "flat", "l", false, "show accounts as a flat list")
	balanceCmd.Flags().BoolVarP(&showZero, "empty", "E", false, "show accounts with zero amount")
	balanceCmd.Flags().BoolVarP(&showBudget, "budget", "B", false, "show budget account balances")
	rootCmd.AddCommand(balanceCmd)
}
