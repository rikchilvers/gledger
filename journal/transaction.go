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

// StateToString converts a TransactionState to a reportable string
func StateToString(state TransactionState) string {
	switch state {
	case UnclearedState:
		return "!"
	case ClearedState:
		return "*"
	default:
		return " "
	}
}

// Transaction holds details about an individual transaction
type Transaction struct {
	Date                    time.Time
	State                   TransactionState
	Payee                   string
	Postings                []*Posting
	postingWithElidedAmount *Posting
	HeaderNote              string   // note in the header
	Notes                   []string // notes under the header
}

// NewTransaction creates a transaction
func NewTransaction() Transaction {
	return Transaction{
		Date:                    time.Time{},
		State:                   NoState,
		Postings:                make([]*Posting, 0, 8),
		postingWithElidedAmount: nil,
		Notes:                   make([]string, 0, 4),
	}
}

func (t Transaction) String() string {
	const dashDateFormat string = "2006-01-02"
	date := t.Date.Format(dashDateFormat)
	rs := fmt.Sprintf("%s %s %s", date, StateToString(t.State), t.Payee)

	if len(t.HeaderNote) > 0 {
		rs = fmt.Sprintf("%s    ; %s", rs, t.HeaderNote)
	}

	for _, n := range t.Notes {
		rs = fmt.Sprintf("%s\n    ; %s", rs, n)
	}

	for _, p := range t.Postings {
		rs = fmt.Sprintf("%s\n    %s", rs, p)
	}

	return fmt.Sprintf("%s\n", rs)
}

// AddNote adds a note to the transaction
func (t *Transaction) AddNote(note string) {
	t.Notes = append(t.Notes, note)
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
		t.postingWithElidedAmount.Amount = NewAmount(c, -sum)
	}

	return nil
}
