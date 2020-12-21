package cmd

import (
	"fmt"
	"sort"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(accountsCmd)
}

var accountsCmd = &cobra.Command{
	Use:          "accounts",
	Aliases:      []string{"acc", "a"},
	Short:        "List all accounts",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, _ []string) {
		aj := newAccountsJournal()
		th := dateCheckedTransactionHandler(aj.transactionHandler)
		if err := parse(th, nil); err != nil {
			fmt.Println(err)
			return
		}
		if err := aj.prepare(); err != nil {
			fmt.Println(err)
			return
		}
		aj.report()
	},
}

type accountsJournal struct {
	uniqueAccounts map[string]bool
	accounts       []string
}

func newAccountsJournal() accountsJournal {
	return accountsJournal{
		uniqueAccounts: make(map[string]bool),
	}
}

func (aj *accountsJournal) transactionHandler(t *journal.Transaction, _ string) error {
	for _, p := range t.Postings {
		aj.uniqueAccounts[p.AccountPath] = true
	}
	return nil
}

func (aj *accountsJournal) prepare() error {
	aj.accounts = make([]string, 0, len(aj.uniqueAccounts))
	for a := range aj.uniqueAccounts {
		aj.accounts = append(aj.accounts, a)
	}

	if len(filters) == 0 {
		sort.Strings(aj.accounts)
		return nil
	}

	filtered := make([]string, 0)
accountsLoop:
	for _, a := range aj.accounts {
		for _, f := range filters {
			if f.MatchesString(a) {
				filtered = append(filtered, a)
				continue accountsLoop
			}
		}
	}

	sort.Strings(filtered)
	aj.accounts = filtered

	return nil
}

func (aj *accountsJournal) report() {
	for _, a := range aj.accounts {
		fmt.Println(a)
	}
}
