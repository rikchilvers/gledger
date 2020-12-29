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
		bp := newBalanceProcessor()
		// TODO: swap periodic transaction handler to bp one
		if err := parse(bp.transactionHandler, bp.journal.AddPeriodicTransaction); err != nil {
			fmt.Println(err)
			return
		}

		prepareBalance(bp.journal)
		report(*bp.journal.Root, flattenTree, collapseOnlyChildren)

		if showBudget {
			fmt.Println("would print budget here")
			// report(*bp.journal.BudgetRoot, flattenTree, collapseOnlyChildren)
		}
	},
}

func init() {
	balanceCmd.Flags().BoolVarP(&flattenTree, "flatten", "F", false, "show accounts as a flat list")
	balanceCmd.Flags().BoolVarP(&showZero, "show-zero", "Z", false, "show accounts with zero amount")
	balanceCmd.Flags().BoolVarP(&collapseOnlyChildren, "collapse", "C", false, "collapse single child accounts into a list")
	rootCmd.AddCommand(balanceCmd)
}

type balanceProcessor struct {
	journal journal.Journal
}

func newBalanceProcessor() balanceProcessor {
	return balanceProcessor{
		journal: journal.NewJournal(),
	}
}

func (bp *balanceProcessor) transactionHandler(t *journal.Transaction, location string) error {
	matchedTransaction, postings, err := checkAgainstFilters(t)
	if err != nil {
		return err
	}

	if !matchedTransaction && len(postings) == 0 {
		return nil
	}

	if matchedTransaction || len(postings) > 0 {
		bp.journal.AddTransaction(t, location)
	}

	for _, p := range postings {
		if err := bp.journal.AddPosting(p); err != nil {
			return err
		}
	}

	return nil
}

// Prepare prepares the Journal for reporting
func prepareBalance(j journal.Journal) {
	if !showZero {
		j.Root.RemoveEmptyChildren()
	}
}
