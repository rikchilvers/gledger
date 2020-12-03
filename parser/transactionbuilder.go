// Package parser handles reading from journal files
package parser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rikchilvers/gledger/journal"
)

//go:generate stringer -type=transactionType
type transactionType int

const (
	normalTransaction transactionType = iota
	periodicTransaction
)

type transactionBuilder struct {
	transaction      interface{}      // the transaction we're building
	transactionType  transactionType  // the type of the transaction being built
	currentPosting   *journal.Posting // the current posting for the transaction
	previousItemType itemType         // the previous item we were given
}

func newTransactionBuilder() transactionBuilder {
	return transactionBuilder{
		transactionType:  normalTransaction,
		previousItemType: -1,
		currentPosting:   nil,
		transaction:      nil,
	}
}

func (tb *transactionBuilder) beginTransaction(t transactionType) {
	switch t {
	case normalTransaction:
		fmt.Println("starting normal transaction")
		tb.transaction = journal.NewTransaction()
		tb.transactionType = normalTransaction
	case periodicTransaction:
		fmt.Println("starting periodic transaction")
		tb.transaction = journal.NewPeriodicTransaction()
		tb.transactionType = periodicTransaction
	}
}

func (tb *transactionBuilder) build(t itemType, content []rune) error {
	switch tb.transactionType {
	case normalTransaction:
		if err := tb.buildNormalTransaction(t, content); err != nil {
			return err
		}
	case periodicTransaction:
		if err := tb.buildPeriodicTransaction(t, content); err != nil {
			return err
		}
	}

	tb.previousItemType = t
	return nil
}

func (tb *transactionBuilder) buildNormalTransaction(t itemType, content []rune) error {
	// Start by casting the transaction
	transaction, ok := tb.transaction.(journal.Transaction)
	if !ok {
		return errors.New("incorrect transaction type")
	}

	switch t {
	case dateItem:
		date, err := parseDate(content)
		if err != nil {
			return err
		}
		transaction.Date = date
	case stateItem:
		if tb.previousItemType != dateItem {
			return fmt.Errorf("expected state but got %s", t)
		}

		switch content[0] {
		case '!':
			transaction.State = journal.UnclearedState
		case '*':
			transaction.State = journal.ClearedState
		default:
			transaction.State = journal.NoState
		}
	case payeeItem:
		if tb.previousItemType != dateItem && tb.previousItemType != stateItem {
			return fmt.Errorf("expected payee but got %s", t)
		}

		// TODO: try to remove necessity of TrimSpace everywhere
		transaction.Payee = strings.TrimSpace(string(content))
	case accountItem:
		if tb.previousItemType != payeeItem &&
			tb.previousItemType != amountItem &&
			tb.previousItemType != accountItem &&
			tb.previousItemType != periodItem {
			return fmt.Errorf("expected account but got %s", t)
		}

		// Skip period transactions
		if tb.transactionType == periodicTransaction {
			break
		}

		// Accounts start a posting, so check if we need to start a new one
		// (When a transaction is started, the current posting is set to nil)
		if tb.currentPosting != nil {
			transaction.AddPosting(tb.currentPosting)
		}
		tb.currentPosting = journal.NewPosting()

		tb.currentPosting.Transaction = &transaction

		tb.currentPosting.AccountPath = string(content)
	case commodityItem:
		if tb.previousItemType != accountItem {
			return fmt.Errorf("expected currency but got %s", t)
		}

		// Skip period transactions
		if tb.transactionType == periodicTransaction {
			break
		}

		if tb.currentPosting.Amount == nil {
			tb.currentPosting.Amount = journal.NewAmount(string(content), 0)
		} else {
			tb.currentPosting.Amount.Commodity = string(content)
		}
	case amountItem:
		if tb.previousItemType != commodityItem && tb.previousItemType != payeeItem {
			return fmt.Errorf("expected amount but got %s", t)
		}

		// Skip period transactions
		if tb.transactionType == periodicTransaction {
			break
		}

		if tb.currentPosting.Amount == nil {
			tb.currentPosting.Amount = journal.NewAmount("", 0)
		}

		amount, err := parseAmount(content)
		if err != nil {
			return fmt.Errorf("error parsing amount: %w", err)
		}
		tb.currentPosting.Amount.Quantity = amount
	}

	// We've modified the transaction so must update it
	tb.transaction = transaction
	return nil
}

func (tb *transactionBuilder) buildPeriodicTransaction(t itemType, content []rune) error {
	transaction, ok := tb.transaction.(journal.PeriodicTransaction)
	if !ok {
		return errors.New("incorrect transaction type")
	}

	switch t {
	case periodItem:
		period, err := parsePeriod(content)
		if err != nil {
			return err
		}
		transaction.Period = period
		fmt.Println("ptransaction has a period")
	default:
		break
	}

	return nil
}

func (tb *transactionBuilder) endTransaction(p Parser) error {
	// If we don't have a transaction, return here
	if tb.transaction == nil {
		return nil
	}

	switch tb.transactionType {
	case normalTransaction:
		return tb.endNormalTransaction(p)
	case periodicTransaction:
		fmt.Println("unimplemented end of periodic transaction")
	}

	return nil
}

func (tb *transactionBuilder) endNormalTransaction(p Parser) error {
	transaction, ok := tb.transaction.(journal.Transaction)
	if !ok {
		return errors.New("incorrect transaction type")
	}

	var err error

	// Make sure we add the last open posting
	if tb.currentPosting != nil {
		err = transaction.AddPosting(tb.currentPosting)
		if err != nil {
			return err
		}
	}

	if err = transaction.Close(); err != nil {
		return err
	}

	if p.transactionHandler != nil {
		if err = p.transactionHandler(&transaction, p.journalFiles[len(p.journalFiles)-1]); err != nil {
			return err
		}
	}

	tb.transaction = nil
	tb.currentPosting = nil
	return nil
}
