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
		bp := newBudgetProcessor()
		if err := parse(bp.transactionHandler, bp.periodicTransactionHandler); err != nil {
			fmt.Println(err)
			return
		}
		prepareBalance(bp.journal)
		bp.report()
	},
}

func init() {
	budgetCmd.Flags().BoolVarP(&flattenTree, "flatten", "F", false, "show accounts as a flat list")
	budgetCmd.Flags().BoolVarP(&showZero, "show-zero", "Z", false, "show accounts with zero amount")
	budgetCmd.Flags().BoolVarP(&collapseOnlyChildren, "collapse", "C", false, "collapse single child accounts into a list")
	rootCmd.AddCommand(budgetCmd)
}

type budgetProcessor struct {
	journal journal.Journal
}

func newBudgetProcessor() budgetProcessor {
	return budgetProcessor{
		journal: journal.NewJournal(journal.ProcessingConfig{
			CalculateBudget: true,
		}),
	}
}

func (bp *budgetProcessor) periodicTransactionHandler(t *journal.PeriodicTransaction, location string) error {
	return bp.journal.AddPeriodicTransaction(t, location)
}

func (bp *budgetProcessor) transactionHandler(t *journal.Transaction, location string) error {
	matchedTransaction, postings, err := checkAgainstFilters(t)
	if err != nil {
		return err
	}

	if !matchedTransaction && len(postings) == 0 {
		return nil
	}

	if matchedTransaction || len(postings) > 0 {
		if err := bp.journal.AddTransaction(t, location); err != nil {
			return err
		}
	}

	for _, p := range postings {
		if err := bp.journal.AddPosting(p); err != nil {
			return err
		}
	}

	return nil
}

func (bp *budgetProcessor) report() {
	for month, budget := range bp.journal.Budget.Months {
		fmt.Println(month)
		report(*budget.EnvelopeRoot, flattenTree, collapseOnlyChildren)
		fmt.Println()
	}
}

func prepareBudget(j *journal.Journal) {
	if !showZero {
		j.Root.RemoveEmptyChildren()

		if showBudget {
			fmt.Println("would remove empty children from budget")
			// j.BudgetRoot.RemoveEmptyChildren()
		}
	}
}
