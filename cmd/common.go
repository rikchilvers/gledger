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
func linkTransaction(r *journal.Account, t *journal.Transaction, filePath string) error {
	// Add postings to accounts
	for _, p := range t.Postings {
		// Wire up the account for the posting
		p.Account = r.FindOrCreateAccount(strings.Split(p.AccountPath, ":"))

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
