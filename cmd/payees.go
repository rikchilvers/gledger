package cmd

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(payeesCmd)
}

var payeesCmd = &cobra.Command{
	Use:          "payees",
	Aliases:      []string{"pay", "p"},
	Short:        "List all payees",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, args []string) {
		pj := newPayeesJournal()
		th := dateCheckedTransactionHandler(pj.transactionHandler)
		if err := parse(th, nil); err != nil {
			fmt.Println(err)
			return
		}
		if err := pj.prepare(args); err != nil {
			fmt.Println(err)
			return
		}
		pj.report()
	},
}

type payeesJournal struct {
	uniquePayees map[string]bool
	payees       []string
}

func newPayeesJournal() payeesJournal {
	return payeesJournal{
		uniquePayees: make(map[string]bool),
	}
}

func (pj *payeesJournal) transactionHandler(t *journal.Transaction, _ string) error {
	pj.uniquePayees[t.Payee] = true
	return nil
}

func (pj *payeesJournal) prepare(args []string) error {
	pj.payees = make([]string, 0, len(pj.uniquePayees))
	for p := range pj.uniquePayees {
		pj.payees = append(pj.payees, p)
	}

	if len(args) == 0 {
		sort.Strings(pj.payees)
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

payeesLoop:
	for _, p := range pj.payees {
		for _, r := range regexes {
			if r.MatchString(p) {
				filtered = append(filtered, p)
				continue payeesLoop
			}
		}
	}

	sort.Strings(filtered)
	pj.payees = filtered

	return nil
}

func (pj *payeesJournal) report() {
	for _, a := range pj.payees {
		fmt.Println(a)
	}
}
