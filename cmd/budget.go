package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

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
		// prepareBalance(bp.journal)
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
	budget journal.Budget
}

func newBudgetProcessor() budgetProcessor {
	return budgetProcessor{
		budget: journal.NewBudget(),
	}
}

func (bp *budgetProcessor) periodicTransactionHandler(pt *journal.PeriodicTransaction, location string) error {
	// return bp.journal.AddPeriodicTransaction(t, location)

	// Convert the periodic transaction to real transactions then link them
	// TODO take time bounds for running periodic transactions
	transactions := pt.Run(time.Time{}, time.Time{})

	// Handle budget allocations differently
	// PeriodicTransaction with no interval are considered budget transactions
	if pt.Period.Interval == journal.PNone {
		for _, p := range transactions[0].Postings {
			// if err := wireUpPosting(j.BudgetRoot, transaction, p); err != nil {
			if err := bp.budget.AddPosting(p, journal.EnvelopePosting); err != nil {
				return err
			}
		}
		return nil
	}

	for _, t := range transactions {
		bp.transactionHandler(&t, location)
	}

	return nil
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
		// bp.journal.AddTransaction(t, location)
		fmt.Println("journal would normally add a transaction here")
	}

	for _, p := range postings {
		switch strings.Split(p.AccountPath, ":")[0] {
		case journal.ExpensesID:
			if err := bp.budget.AddPosting(p, journal.ExpensePosting); err != nil {
				return err
			}
		case journal.IncomeID:
			if err := bp.budget.AddPosting(p, journal.IncomePosting); err != nil {
				return err
			}
		}
	}

	return nil
}

func (bp *budgetProcessor) report() {
	// sort the months
	months := make([]time.Time, 0, len(bp.budget.Months))
	for month := range bp.budget.Months {
		months = append(months, month)
	}
	sort.Slice(months, func(i, j int) bool {
		return months[i].Before(months[j])
	})

	for _, month := range months {
		budget := bp.budget.Months[month]
		fmt.Println(month)

		// Print the funds
		fmt.Printf("%20s  %s\n", budget.Income.Amount.DisplayableQuantity(true), budget.Income.Name)
		fmt.Printf("%20s  %s\n", budget.ExpenseRoot.Amount.DisplayableQuantity(true), budget.ExpenseRoot.Name)

		// Print the envelopes
		for _, cn := range budget.EnvelopeRoot.SortedChildNames() {
			envelopeAccount, found := budget.EnvelopeRoot.Children[cn]
			if !found {
				fmt.Println("didn't find envelope account", cn)
				return
			}

			expenseAccount, found := budget.ExpenseRoot.Children[cn]
			if !found {
				fmt.Println("didn't find expense account", cn)
				return
			}

			amount := envelopeAccount.Amount
			if err := amount.Add(expenseAccount.Amount); err != nil {
				panic(err)
			}

			fmt.Printf("%20s  %20s  %20s  %s\n", expenseAccount.Amount, envelopeAccount.Amount, amount.DisplayableQuantity(true), cn)
		}

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
