package main

type account struct {
	// The user visible name for the account
	name string
	// Parent account (e.g. Parent:Child)
	parent *account
	// All children of the account
	quantity     int64
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
