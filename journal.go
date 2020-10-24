package main

type journal struct {
	rootAccount  *account
	transactions []*transaction
}

func newJournal() *journal {
	return &journal{
		rootAccount:  newAccount("root"),
		transactions: make([]*transaction, 0, 1024),
	}
}

func (j *journal) addTransaction(t *transaction) {
	j.transactions = append(j.transactions, t)
}
