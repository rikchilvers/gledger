package main

import (
	"log"
	"time"
)

type TransactionState int

const (
	NoState TransactionState = iota
	Uncleared
	Cleared
)

type Transaction struct {
	date                      time.Time
	state                     TransactionState
	payee                     string
	postingsWithElidedAmounts int
	postings                  []Posting
}

func newTransaction() Transaction {
	return Transaction{}
}

func (t *Transaction) addPosting(p Posting) error {
	if _, ok := p.amount.(float32); ok {
		t.postingsWithElidedAmounts++
		if t.postingsWithElidedAmounts > 1 {
			log.Fatalln("Cannot have more than one posting with an elided amount")
		}
	}
	t.postings = append(t.postings, p)

	return nil
}

func (t *Transaction) close() error {
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
