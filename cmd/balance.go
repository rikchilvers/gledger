package cmd

import (
	"fmt"
	"regexp"

	"github.com/rikchilvers/gledger/journal"
	"github.com/rikchilvers/gledger/reporting"
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
		config := journal.ProcessingConfig{
			CalculateBudget: showBudget,
		}
		journal := journal.NewJournal(config)

		th := dateCheckedTransactionHandler(journal.AddTransaction)
		if err := parse(th, journal.AddPeriodicTransaction); err != nil {
			fmt.Println(err)
			return
		}

		prepareBalance(journal)
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

// Prepare prepares the Journal for reporting
func prepareBalance(j journal.Journal) {
	// Filter output with account name filters
	if len(filters) > 0 {
		j.Root.RemoveChildren(func(a journal.Account) bool {
			for _, f := range filters {
				if f.FilterType != reporting.AccountNameFilter {
					continue
				}

				fmt.Printf("matching against %25s (%s)\n", a.Name, a.Path)

				if f.MatchesString(a.Path) {
					return true
				}
			}

			return false
		})
	}

	if !showZero {
		j.Root.RemoveEmptyChildren()

		// showBudget is the same as journal.config.CalculateBudget
		// TODO rename one of them
		if showBudget {
			j.BudgetRoot.RemoveEmptyChildren()
		}
	}
}

func stringMatchesRegex(input string, args []string) (bool, error) {
	if len(args) == 0 {
		return true, nil
	}

	for _, arg := range args {
		if !reporting.ContainsUppercase(arg) {
			arg = "(?i)" + arg
		}
		regex, err := regexp.Compile(arg)
		if err != nil {
			return false, err
		}

		if regex.MatchString(input) {
			return true, nil
		}
	}

	return false, nil
}
