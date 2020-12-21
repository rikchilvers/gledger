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
	Run: func(_ *cobra.Command, args []string) {
		config := journal.ProcessingConfig{
			CalculateBudget: showBudget,
		}
		journal := journal.NewJournal(config)
		th := dateCheckedTransactionHandler(journal.AddTransaction)
		if err := parse(th, journal.AddPeriodicTransaction); err != nil {
			fmt.Println(err)
			return
		}
		prepareBalance(journal, args)
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
func prepareBalance(j journal.Journal, args []string) {
	matchedAccounts := j.Root.FindAccounts(func(a journal.Account) bool {
		matches, _ := stringMatchesRegex(a.Name, args)
		return !matches
	})

	for _, a := range matchedAccounts {
		fmt.Println("unlinking", a.Name)
		a.Unlink()
	}

	if !showZero {
		j.Root.RemoveEmptyChildren()

		// showBudget is the same as journal.config.CalculateBudget
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
