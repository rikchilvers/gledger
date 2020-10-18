package main

import (
	"fmt"
	"log"
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
	date                      time.Time
	state                     transactionState
	payee                     string
	postingsWithElidedAmounts int
	postings                  []*posting
}

func newTransaction() *transaction {
	return &transaction{
		date:                      time.Time{},
		state:                     tNoState,
		payee:                     "",
		postingsWithElidedAmounts: 0,
		postings:                  make([]*posting, 0, 8),
	}
}

func (t transaction) String() string {
	return fmt.Sprintf("Transaction:\n\t%s\n\t%s\n\t%s\n\t%d postings (%d)", t.date, t.state.String(), t.payee, len(t.postings), t.postingsWithElidedAmounts)
}

func (t *transaction) addPosting(p *posting) error {
	if p.amount == nil {
		t.postingsWithElidedAmounts++
		if t.postingsWithElidedAmounts > 1 {
			log.Fatalln("Cannot have more than one posting with an elided amount")
		}
	}
	t.postings = append(t.postings, p)

	return nil
}

func (t *transaction) close() error {
	sum := int64(0)
	for _, p := range t.postings {
		if p.amount == nil {
			continue
		} else {
			sum += p.amount.quantity
		}
	}
	return nil
}
