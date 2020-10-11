package main

import "time"

type TransactionState int

const (
	Cleared TransactionState = iota
	Uncleared
	NoState
)

type Transaction struct {
	date     time.Time
	state    TransactionState
	payee    string
	postings []Posting
}

func newTransaction() Transaction {
	return Transaction{}
}

type Posting struct {
	account  string
	currency string
	amount   float32
}

func newPosting(account, currency string, amount float32) Posting {
	return Posting{
		account,
		currency,
		amount,
	}
}
