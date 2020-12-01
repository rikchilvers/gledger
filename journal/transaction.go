package journal

import (
	"errors"
	"fmt"
	"time"
)

//go:generate stringer -type=TransactionState
// TransactionState represents the state a Transaction can be in
type TransactionState int

// State a Transaction can be in
const (
	NoState TransactionState = iota
	UnclearedState
	ClearedState
)

// Transaction holds details about an individual transaction
type Transaction struct {
	Date                    time.Time
	State                   TransactionState
	Payee                   string
	Postings                []*Posting
	postingWithElidedAmount *Posting
}

// NewTransaction creates a transaction
func NewTransaction() Transaction {
	return Transaction{
		Date:                    time.Time{},
		State:                   NoState,
		Payee:                   "",
		Postings:                make([]*Posting, 0, 8),
		postingWithElidedAmount: nil,
	}
}

func (t Transaction) String() string {
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

// AddPosting adds a posting to the Transaction (ensuring there is only one with an elided amount)
func (t *Transaction) AddPosting(p *Posting) error {
	if p.Amount == nil {
		if t.postingWithElidedAmount != nil {
			return errors.New("cannot have more than one posting with an elided amount")
		}
		t.postingWithElidedAmount = p
	}

	// TODO: during parsing, check amounts with commodities cannot be created without amounts

	t.Postings = append(t.Postings, p)

	return nil
}

// Close ensures the transaction balances (assigning an amount to an elided posting as necessary)
func (t *Transaction) Close() error {
	// Check the transaction balances
	sum := int64(0)
	c := ""
	for _, p := range t.Postings {
		if p.Amount == nil {
			continue
		}
		// if there is a commodity, take a note of it
		if len(p.Amount.Commodity) > 0 {
			c = p.Amount.Commodity
		}
		sum += p.Amount.Quantity
	}

	if sum != 0 {
		if t.postingWithElidedAmount == nil {
			return errors.New("transaction does not balance")
		}

		// NB: setting the commodity like this will not work with multiple currencies
		// see above comment too
		t.postingWithElidedAmount.Amount = NewAmount(c, -sum)
	}

	return nil
}
