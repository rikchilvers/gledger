// Package cmd handles the cli
package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/rikchilvers/gledger/journal"
	"github.com/rikchilvers/gledger/parser"
)

func parse(th parser.TransactionHandler, ph parser.PeriodicTransactionHandler) error {
	if rootJournalPath == "" {
		// TODO: use viper to read env variable
		return errors.New("no root journal path provided")
	}

	// Open the file
	file, err := os.Open(rootJournalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	p := parser.NewParser(th, ph)
	if err := p.Parse(file, rootJournalPath); err != nil {
		return err
	}
	return nil
}

func report(account *journal.Account, flattenTree bool) {
	prepender := func(a journal.Account) string {
		return fmt.Sprintf("%20s  ", a.Amount.DisplayableQuantity(true))
	}

	if flattenTree {
		flattened := account.FlattenedTree(prepender)
		fmt.Println(flattened)
	} else {
		tree := account.Tree(prepender)
		fmt.Println(tree)
	}

	// 20x '-' because that is how wide we format the amount to be
	fmt.Println("--------------------")

	// Print the root account's value
	fmt.Printf("%20s\n", account.Amount.DisplayableQuantity(false))
}

// linkTransaction builds the account tree
func linkTransaction(root *journal.Account, transaction *journal.Transaction, _ string) error {
	for _, p := range transaction.Postings {
		// Use the parsed account path to find or create the account
		if p.Account == nil {
			p.Account = root.FindOrCreateAccount(strings.Split(p.AccountPath, ":"))
		}

		// Add postings to accounts
		p.Account.Postings = append(p.Account.Postings, p)

		// Add the transaction to the account
		p.Account.Transactions = append(p.Account.Transactions, transaction)

		// Add the posting's amount to the account and all its ancestors
		if err := p.Account.WalkAncestors(func(a *journal.Account) error {
			if err := a.Amount.Add(p.Amount); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
