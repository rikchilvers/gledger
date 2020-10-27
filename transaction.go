package main

import (
	"errors"
	"fmt"
	"time"
)

//go:generate stringer -type=transactionState
type transactionState int

const (
	tNoState transactionState = iota
	tUncleared
	tCleared
)

type transaction struct {
	date                    time.Time
	state                   transactionState
	payee                   string
	postingWithElidedAmount *posting
	postings                []*posting
}

func newTransaction() *transaction {
	return &transaction{
		date:                    time.Time{},
		state:                   tNoState,
		payee:                   "",
		postingWithElidedAmount: nil,
		postings:                make([]*posting, 0, 8),
	}
}

func (t transaction) String() string {
	// return fmt.Sprintf("Transaction:\n\t%s\n\t%s\n\t%s\n\t%d postings (%v)", t.date, t.state.String(), t.payee, len(t.postings), t.postingWithElidedAmount != nil)
	ts := fmt.Sprintf(`Transaction:
	%s
	%s
	%s
	%d postings (%v)`, t.date, t.state.String(), t.payee, len(t.postings), t.postingWithElidedAmount != nil)
	for _, p := range t.postings {
		ts = fmt.Sprintf("%s\n\t\t%s", ts, p)
	}
	return ts
}

func (t *transaction) addPosting(p *posting) error {
	if p.amount == nil {
		if t.postingWithElidedAmount != nil {
			return errors.New("Cannot have more than one posting with an elided amount")
		}
		t.postingWithElidedAmount = p
	}

	// TODO: during parsing, check amounts with commodities cannot be created without amounts

	t.postings = append(t.postings, p)

	return nil
}

func (t *transaction) close() error {
	// Check the transaction balances
	sum := int64(0)
	for _, p := range t.postings {
		if p.amount == nil {
			continue
		} else {
			sum += p.amount.quantity
		}
	}

	if sum != 0 {
		if t.postingWithElidedAmount == nil {
			return errors.New("transaction does not balance")
		}
		t.postingWithElidedAmount.amount = newAmount(-sum)
	}

	return nil
}
