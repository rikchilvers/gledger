package cmd

import (
	"fmt"
	"sort"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(printCmd)
}

var printCmd = &cobra.Command{
	Use:          "print",
	Aliases:      []string{"p"},
	Short:        "Shows transaction entries, sorted by date",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, _ []string) {
		pj := newPrintJournal()
		th := dateCheckedFilteringTransactionHandler(pj.transactionHandler)
		if err := parse(th, nil); err != nil {
			fmt.Println(err)
			return
		}
		pj.prepare()
		pj.report()
	},
}

type printJournal struct {
	transactions []*journal.Transaction
}

func newPrintJournal() printJournal {
	return printJournal{
		transactions: make([]*journal.Transaction, 0, 2056),
	}
}

func (pj *printJournal) transactionHandler(t *journal.Transaction, _ string) error {
	pj.transactions = append(pj.transactions, t)
	return nil
}

func (pj *printJournal) prepare() {
	sort.Slice(pj.transactions, func(i, j int) bool {
		return pj.transactions[i].Date.Before(pj.transactions[j].Date)
	})
}

func (pj *printJournal) report() {
	for _, t := range pj.transactions {
		fmt.Println(t)
	}
}
