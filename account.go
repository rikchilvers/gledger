package main

type account struct {
	name         string
	quantity     int64
	parent       *account
	children     []*account
	postings     []*posting
	transactions []*transaction
}

func newAccount(name string) *account {
	return &account{
		name:         name,
		quantity:     0,
		parent:       nil,
		children:     make([]*account, 0, 16),
		postings:     make([]*posting, 0, 2048),
		transactions: make([]*transaction, 0, 1024),
	}
}
