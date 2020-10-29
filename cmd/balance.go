package cmd

import (
	"fmt"
	"strings"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(balanceCmd)
}

var balanceCmd = &cobra.Command{
	Use:          "balance",
	Aliases:      []string{"bal", "b"},
	Short:        "Shows accounts and their balances",
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		jb := newJournalBalance()
		if err := parse(jb.analyseTransaction); err != nil {
			fmt.Println(err)
			return
		}
		jb.report()
	},
}

type journalBalance struct {
	rootAccount *journal.Account
}

func newJournalBalance() journalBalance {
	return journalBalance{
		rootAccount: journal.NewAccount("root"),
	}
}

func (jb *journalBalance) analyseTransaction(t *journal.Transaction, p string) error {
	// Add postings to accounts
	for _, p := range t.Postings {
		// Wire up the account for the posting
		p.Account = jb.rootAccount.FindOrCreateAccount(strings.Split(p.AccountPath, ":"))

		// Apply amount to each the account and all its ancestors
		account := p.Account
		for {
			if account.Parent == nil {
				break
			}
			account.Amount.Quantity += p.Amount.Quantity
			account = account.Parent
		}

		// Tie up references
		p.Account.Postings = append(p.Account.Postings, p)
		p.Account.Transactions = append(p.Account.Transactions, t)
	}

	return nil
}

func (jb journalBalance) report() {
	printAccountsAndQuantities(*jb.rootAccount, 0)
	// Print the root account's value
	fmt.Println("--------------------") // 20 is how wide we format the amount to be
	fmt.Printf("%20s\n", jb.rootAccount.Amount.DisplayableQuantity(false))
}

func printAccountsAndQuantities(a journal.Account, depth int) {
	const tabWidth int = 2

	if a.Name == "root" {
		for _, child := range a.Children {
			printAccountsAndQuantities(*child, depth+1)
		}
		return
	}

	spaces := ""
	for i := 0; i < depth*tabWidth; i++ {
		spaces = fmt.Sprintf(" %s", spaces)
	}
	nameAndQuantity := fmt.Sprintf("%20s%s%s", a.Amount.DisplayableQuantity(true), spaces, a.Name)
	fmt.Println(nameAndQuantity)

	for _, child := range a.Children {
		printAccountsAndQuantities(*child, depth+1)
	}
}
