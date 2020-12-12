// Package parser handles reading from journal files
package parser

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
)

//go:generate stringer -type=transactionType
type transactionType int

const (
	normalTransaction transactionType = iota
	periodicTransaction
)

type transactionBuilder struct {
	transactionType     transactionType              // the type of the transaction being built
	transaction         *journal.Transaction         // the transaction we're building
	periodicTransaction *journal.PeriodicTransaction // the periodic transaction we're building
	currentPosting      *journal.Posting             // the current posting for the transaction
	previousItemType    itemType                     // the previous item we were given
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
		transaction := journal.NewTransaction()
		tb.transaction = &transaction
	case periodicTransaction:
		transaction := journal.NewPeriodicTransaction()
		tb.periodicTransaction = &transaction
	}
}

func (tb *transactionBuilder) build(t itemType, content []rune) error {
	switch tb.transactionType {
	case normalTransaction:
		if err := tb.buildNormalTransaction(tb.transaction, t, content); err != nil {
			return err
		}
	case periodicTransaction:
		if err := tb.buildPeriodicTransaction(tb.periodicTransaction, t, content); err != nil {
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

		t.Payee = string(content)
	case accountItem:
		if tb.previousItemType != payeeItem &&
			tb.previousItemType != amountItem &&
			tb.previousItemType != accountItem &&
			tb.previousItemType != periodItem {
			return fmt.Errorf("expected account but got %s", item)
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

		if tb.currentPosting.Amount == nil {
			tb.currentPosting.Amount = journal.NewAmount(string(content), 0)
		} else {
			tb.currentPosting.Amount.Commodity = string(content)
		}
	case amountItem:
		if tb.previousItemType != commodityItem && tb.previousItemType != payeeItem {
			return fmt.Errorf("expected amount but got %s", item)
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
		// fmt.Println("deferring to buildNormalTransaction")
		return tb.buildNormalTransaction(&t.Transaction, i, content)
	}

	return nil
}

func (tb *transactionBuilder) endTransaction(p Parser) error {
	switch tb.transactionType {
	case normalTransaction:
		// If there is no transaction, bail here
		if tb.transaction == nil {
			return nil
		}

		if err := tb.endNormalTransaction(tb.transaction, p); err != nil {
			return err
		}

		if p.transactionHandler != nil {
			if err := p.transactionHandler(tb.transaction, p.journalFiles[len(p.journalFiles)-1]); err != nil {
				return err
			}
		}

		tb.transaction = nil
	case periodicTransaction:
		// If there is no transaction, bail here
		if tb.periodicTransaction == nil {
			return nil
		}

		if err := tb.endNormalTransaction(&tb.periodicTransaction.Transaction, p); err != nil {
			return err
		}

		if p.periodicTransactionHandler != nil {
			if err := p.periodicTransactionHandler(tb.periodicTransaction, p.journalFiles[len(p.journalFiles)-1]); err != nil {
				return err
			}
		}

		tb.periodicTransaction = nil
	}

	tb.currentPosting = nil
	return nil
}

func (tb *transactionBuilder) endNormalTransaction(t *journal.Transaction, p Parser) error {
	// Make sure we add the last open posting
	if tb.currentPosting != nil {
		if err := t.AddPosting(tb.currentPosting); err != nil {
			return err
		}
	}

	// Before we close a budget transaction, we need to add the 'To Be Budgeted' account
	if tb.transactionType == periodicTransaction && tb.periodicTransaction.Period.Interval == journal.PNone {
		hasBudgetSource := false
		for _, p := range t.Postings {
			if p.AccountPath == journal.ToBeBudgetedID {
				hasBudgetSource = true
				break
			}
		}

		if !hasBudgetSource {
			source := journal.NewPosting()
			source.Transaction = t
			source.AccountPath = journal.ToBeBudgetedID
			t.AddPosting(source)
		}
	}

	if err := t.Close(); err != nil {
		return err
	}

	return nil
}
