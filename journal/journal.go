package journal

import "strings"

type Journal struct {
	rootAccount  *Account
	transactions []*Transaction
}

func NewJournal() *Journal {
	return &Journal{
		rootAccount:  NewAccount("root"),
		transactions: make([]*Transaction, 0, 1024),
	}
}

func (j *Journal) AddTransaction(t *Transaction) {
	j.transactions = append(j.transactions, t)

	// Add postings to accounts
	for _, p := range t.Postings {
		// Wire up the account for the posting
		p.Account = j.rootAccount.findOrCreateAccount(strings.Split(p.AccountPath, ":"))

		// Apply amount to each the account and all its ancestors
		account := p.Account
		for {
			if account.Parent == nil {
				break
			}
			account.Amount.Quantity += p.Amount.Quantity
			account = account.Parent
		}

		// Tie up references
		p.Account.Postings = append(p.Account.Postings, p)
		p.Account.Transactions = append(p.Account.Transactions, t)
	}
}
