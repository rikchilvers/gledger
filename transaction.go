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
	postings                  []posting
}

func newTransaction() *transaction {
	return &transaction{}
}

func (t transaction) String() string {
	return fmt.Sprintf("Transaction:\n\t%s\n\t%s\n\t%s\n\t%d postings (%d)", t.date, t.state.String(), t.payee, len(t.postings), t.postingsWithElidedAmounts)
}

func (t *transaction) addPosting(p posting) error {
	if _, ok := p.amount.(float32); ok {
		t.postingsWithElidedAmounts++
		if t.postingsWithElidedAmounts > 1 {
			log.Fatalln("Cannot have more than one posting with an elided amount")
		}
	}
	t.postings = append(t.postings, p)

	return nil
}

func (t *transaction) close() error {
	sum := float32(0)
	for _, p := range t.postings {
		if amount, ok := p.amount.(float32); ok {
			sum += amount
		} else {
			continue
		}
	}
	return nil
}
