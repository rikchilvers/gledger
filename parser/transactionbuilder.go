// Package parser handles reading from journal files
package parser

import (
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
	transactionType     transactionType             // the type of the transaction being built
	transaction         journal.Transaction         // the transaction we're building
	periodicTransaction journal.PeriodicTransaction // the periodic transaction we're building
	currentPosting      *journal.Posting            // the current posting for the transaction
	previousItemType    itemType                    // the previous item we were given
}

func newTransactionBuilder() transactionBuilder {
	return transactionBuilder{
		transactionType:  normalTransaction,
		previousItemType: -1,
		currentPosting:   nil,
	}
}

func (tb *transactionBuilder) beginTransaction(t transactionType) {
	tb.transactionType = t
	switch t {
	case normalTransaction:
		tb.transaction = journal.NewTransaction()
	case periodicTransaction:
		tb.periodicTransaction = journal.NewPeriodicTransaction()
	}
}

func (tb *transactionBuilder) build(t itemType, content []rune) error {
	switch tb.transactionType {
	case normalTransaction:
		if err := tb.buildNormalTransaction(&tb.transaction, t, content); err != nil {
			return err
		}
	case periodicTransaction:
		if err := tb.buildPeriodicTransaction(&tb.periodicTransaction, t, content); err != nil {
			return err
		}
	}

	tb.previousItemType = t
	return nil
}

func (tb *transactionBuilder) buildNormalTransaction(t *journal.Transaction, item itemType, content []rune) error {
	switch item {
	case dateItem:
		date, err := parseDate(content)
		if err != nil {
			return err
		}
		t.Date = date
	case stateItem:
		if tb.previousItemType != dateItem {
			return fmt.Errorf("expected state but got %s", item)
		}

		switch content[0] {
		case '!':
			t.State = journal.UnclearedState
		case '*':
			t.State = journal.ClearedState
		default:
			t.State = journal.NoState
		}
	case payeeItem:
		if tb.previousItemType != dateItem && tb.previousItemType != stateItem {
			return fmt.Errorf("expected payee but got %s", item)
		}

		// TODO: try to remove necessity of TrimSpace everywhere
		t.Payee = strings.TrimSpace(string(content))
	case accountItem:
		if tb.previousItemType != payeeItem &&
			tb.previousItemType != amountItem &&
			tb.previousItemType != accountItem &&
			tb.previousItemType != periodItem {
			return fmt.Errorf("expected account but got %s", item)
		}

		// Skip period transactions
		if tb.transactionType == periodicTransaction {
			break
		}

		// Accounts start a posting, so check if we need to start a new one
		// (When a transaction is started, the current posting is set to nil)
		if tb.currentPosting != nil {
			t.AddPosting(tb.currentPosting)
		}
		tb.currentPosting = journal.NewPosting()

		tb.currentPosting.Transaction = t

		tb.currentPosting.AccountPath = string(content)
	case commodityItem:
		if tb.previousItemType != accountItem {
			return fmt.Errorf("expected currency but got %s", item)
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
			return fmt.Errorf("expected amount but got %s", item)
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

	return nil
}

func (tb *transactionBuilder) buildPeriodicTransaction(t *journal.PeriodicTransaction, i itemType, content []rune) error {
	switch i {
	case periodItem:
		period, err := parsePeriod(content)
		if err != nil {
			return err
		}
		t.Period = period
	default:
		// In all other cases, we just want to build the normal transaction
		return tb.buildNormalTransaction(&t.Transaction, i, content)
	}

	return nil
}

func (tb *transactionBuilder) endTransaction(p Parser) error {
	// If the transaction hasn't been modified, end here
	if tb.transaction.Date.IsZero() {
		return nil
	}

	switch tb.transactionType {
	case normalTransaction:
		return tb.endNormalTransaction(&tb.transaction, p)
	case periodicTransaction:
		return tb.endNormalTransaction(&tb.periodicTransaction.Transaction, p)
	}

	return nil
}

func (tb *transactionBuilder) endNormalTransaction(t *journal.Transaction, p Parser) error {
	var err error

	// Make sure we add the last open posting
	if tb.currentPosting != nil {
		err = t.AddPosting(tb.currentPosting)
		if err != nil {
			return err
		}
	}

	if err = t.Close(); err != nil {
		return err
	}

	if p.transactionHandler != nil {
		if err = p.transactionHandler(t, p.journalFiles[len(p.journalFiles)-1]); err != nil {
			return err
		}
	}

	tb.transaction = journal.NewTransaction()
	tb.currentPosting = nil
	return nil
}
