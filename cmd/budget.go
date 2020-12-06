package cmd

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

var budgetCmd = &cobra.Command{
	Use:          "balance",
	Aliases:      []string{"bal", "b"},
	Short:        "Shows accounts and their balances",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		l := newLedger()
		if err := parse(l.transactionHandler, l.periodicTransactionHandler); err != nil {
			fmt.Println(err)
			return
		}
		l.prepare()
		l.report()
	},
}

type ledger struct {
	rootAccount   *journal.Account
	budgetAccount *journal.Account
}

func newLedger() ledger {
	return ledger{
		rootAccount:   journal.NewAccount(journal.RootID),
		budgetAccount: journal.NewAccount("_budget_"),
	}
}

func (l *ledger) transactionHandler(t *journal.Transaction, path string) error {
	if err := linkTransaction(l.rootAccount, t, path); err != nil {
		return err
	}
	return nil
}

func (l *ledger) periodicTransactionHandler(t *journal.PeriodicTransaction, path string) error {
	return nil
}

func (l *ledger) prepare() {}
func (l *ledger) report()  {}
