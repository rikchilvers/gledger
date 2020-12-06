// Package cmd handles the cli
package cmd

import (
	"errors"
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
