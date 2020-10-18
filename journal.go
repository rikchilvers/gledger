package main

type journal struct {
	accounts     []*account
	transactions []*transaction
}

func newJournal() *journal {
	return &journal{
		accounts:     make([]*account, 0, 256),
		transactions: make([]*transaction, 0, 1024),
	}
}
