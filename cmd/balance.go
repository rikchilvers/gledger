package cmd

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

var (
	flattenTree          bool
	collapseOnlyChildren bool
	showZero             bool
	showBudget           bool
)

var balanceCmd = &cobra.Command{
	Use:          "balance",
	Aliases:      []string{"bal", "b"},
	Short:        "Shows accounts and their balances",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, _ []string) {
		config := journal.JournalConfig{
			CalculateBudget: showBudget,
		}
		journal := journal.NewJournal(config)
		if err := parse(journal.AddTransaction, journal.AddPeriodicTransaction); err != nil {
			fmt.Println(err)
			return
		}
		journal.Prepare(showZero)
		report(*journal.Root, flattenTree, collapseOnlyChildren)

		if showBudget {
			fmt.Println("")
			report(*journal.BudgetRoot, flattenTree, collapseOnlyChildren)
		}
	},
}

func init() {
	balanceCmd.Flags().BoolVarP(&flattenTree, "flatten", "F", false, "show accounts as a flat list")
	balanceCmd.Flags().BoolVarP(&showZero, "show-zero", "Z", false, "show accounts with zero amount")
	balanceCmd.Flags().BoolVarP(&showBudget, "show-budget", "B", false, "show budget account balances")
	balanceCmd.Flags().BoolVarP(&collapseOnlyChildren, "collapse", "C", false, "collapse single child accounts into a list")
	rootCmd.AddCommand(balanceCmd)
}
