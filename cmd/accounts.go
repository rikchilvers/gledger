package cmd

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(accountsCmd)
}

var accountsCmd = &cobra.Command{
	Use:          "accounts",
	Aliases:      []string{"a"},
	Short:        "List all accounts",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, args []string) {
		aj := newAccountsJournal()
		th := dateCheckedTransactionHandler(aj.transactionHandler)
		if err := parse(th, nil); err != nil {
			fmt.Println(err)
			return
		}
		if err := aj.prepare(args); err != nil {
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

func (aj *accountsJournal) prepare(args []string) error {
	aj.accounts = make([]string, 0, len(aj.uniqueAccounts))
	for a := range aj.uniqueAccounts {
		aj.accounts = append(aj.accounts, a)
	}

	if len(args) == 0 {
		sort.Strings(aj.accounts)
		return nil
	}

	filtered := make([]string, 0)
	regexes := make([]*regexp.Regexp, 0, len(args))
	for _, arg := range args {
		if !containsUppercase(arg) {
			arg = "(?i)" + arg
		}
		regex, err := regexp.Compile(arg)
		if err != nil {
			return err
		}
		regexes = append(regexes, regex)
	}

accountsLoop:
	for _, a := range aj.accounts {
		for _, r := range regexes {
			if r.MatchString(a) {
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
