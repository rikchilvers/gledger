package cmd

import (
	"errors"
	"strings"

	"github.com/rikchilvers/gledger/journal"
	"github.com/rikchilvers/gledger/parser"
)

func parse(handler parser.TransactionHandler) error {
	if rootJournalPath == "" {
		// TODO: use viper to read env variable
		return errors.New("No root journal path provided")
	}

	p := parser.NewParser(handler)
	if err := p.Parse(rootJournalPath); err != nil {
		return err
	}
	return nil
}

// linkTransaction builds the account tree
func linkTransaction(r *journal.Account, t *journal.Transaction, fp string) error {
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
