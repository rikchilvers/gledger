package cmd

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

var (
	flattenTree bool
	showZero    bool
)

var balanceCmd = &cobra.Command{
	Use:          "balance",
	Aliases:      []string{"bal", "b"},
	Short:        "Shows accounts and their balances",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		jb := newJournalBalance()
		if err := parse(jb.transactionHandler, nil); err != nil {
			fmt.Println(err)
			return
		}
		jb.prepare()
		jb.report()
	},
}

func init() {
	balanceCmd.Flags().BoolVarP(&flattenTree, "flat", "l", false, "show accounts as a flat list")
	balanceCmd.Flags().BoolVarP(&showZero, "empty", "E", false, "show accounts with zero amount")
	rootCmd.AddCommand(balanceCmd)
}

type journalBalance struct {
	rootAccount *journal.Account
}

func newJournalBalance() journalBalance {
	return journalBalance{
		rootAccount: journal.NewAccount(journal.RootID),
	}
}

func (jb *journalBalance) transactionHandler(t *journal.Transaction, path string) error {
	// Defer to the common transaction linker
	if err := linkTransaction(jb.rootAccount, t, path); err != nil {
		return err
	}
	return nil
}

func (jb journalBalance) prepare() {
	if !showZero {
		matcher := func(a journal.Account) bool {
			return a.Amount.Quantity == 0
		}
		matching := jb.rootAccount.FindAccounts(matcher)
		for _, m := range matching {
			if m.Name == journal.RootID {
				continue
			}
			// remove the account from it's parent
			delete(m.Parent.Children, m.Name)
			m.Parent = nil
		}
	}
}

func (jb journalBalance) report() {
	prepender := func(a journal.Account) string {
		return fmt.Sprintf("%20s  ", a.Amount.DisplayableQuantity(true))
	}

	if flattenTree {
		flattened := jb.rootAccount.FlattenedTree(prepender)
		fmt.Println(flattened)
	} else {
		tree := jb.rootAccount.Tree(prepender)
		fmt.Println(tree)
	}

	// 20 - because that is how wide we format the amount to be
	fmt.Println("--------------------")

	// Print the root account's value
	fmt.Printf("%20s\n", jb.rootAccount.Amount.DisplayableQuantity(false))
}
