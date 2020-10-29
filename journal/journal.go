package journal

import "strings"

type journal struct {
	rootAccount  *Account
	transactions []*Transaction
}

func newJournal() *journal {
	return &journal{
		rootAccount:  NewAccount("root"),
		transactions: make([]*Transaction, 0, 1024),
	}
}

func (j *journal) addTransaction(t *Transaction) {
	j.transactions = append(j.transactions, t)

	// Add postings to accounts
	for _, p := range t.Postings {
		// Wire up the account for the posting
		p.Account = j.rootAccount.FindOrCreateAccount(strings.Split(p.AccountPath, ":"))

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
