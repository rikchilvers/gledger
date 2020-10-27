package journal

import (
	"errors"
	"fmt"
	"time"
)

//go:generate stringer -type=TransactionState
type TransactionState int

const (
	NoState TransactionState = iota
	UnclearedState
	ClearedState
)

type Transaction struct {
	Date                    time.Time
	State                   TransactionState
	Payee                   string
	Postings                []*Posting
	postingWithElidedAmount *Posting
}

func NewTransaction() *Transaction {
	return &Transaction{
		Date:                    time.Time{},
		State:                   NoState,
		Payee:                   "",
		Postings:                make([]*Posting, 0, 8),
		postingWithElidedAmount: nil,
	}
}

func (t Transaction) String() string {
	// return fmt.Sprintf("Transaction:\n\t%s\n\t%s\n\t%s\n\t%d postings (%v)", t.date, t.state.String(), t.payee, len(t.postings), t.postingWithElidedAmount != nil)
	ts := fmt.Sprintf(`Transaction:
	%s
	%s
	%s
	%d postings (%v)`, t.Date, t.State.String(), t.Payee, len(t.Postings), t.postingWithElidedAmount != nil)
	for _, p := range t.Postings {
		ts = fmt.Sprintf("%s\n\t\t%s", ts, p)
	}
	return ts
}

func (t *Transaction) AddPosting(p *Posting) error {
	if p.Amount == nil {
		if t.postingWithElidedAmount != nil {
			return errors.New("Cannot have more than one posting with an elided amount")
		}
		t.postingWithElidedAmount = p
	}

	// TODO: during parsing, check amounts with commodities cannot be created without amounts

	t.Postings = append(t.Postings, p)

	return nil
}

func (t *Transaction) Close() error {
	// Check the transaction balances
	sum := int64(0)
	for _, p := range t.Postings {
		if p.Amount == nil {
			continue
		} else {
			sum += p.Amount.Quantity
		}
	}

	if sum != 0 {
		if t.postingWithElidedAmount == nil {
			return errors.New("transaction does not balance")
		}
		t.postingWithElidedAmount.Amount = NewAmount(-sum)
	}

	return nil
}
