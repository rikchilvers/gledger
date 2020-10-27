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

	// Add postings to accounts
	for _, p := range t.postings {
		// Wire up the account for the posting
		p.account = j.rootAccount.findOrCreateAccount(p.accountPath)

		// Apply amount to each the account and all its ancestors
		account := p.account
		for {
			if account.parent == nil {
				break
			}
			account.amount.quantity += p.amount.quantity
			account = account.parent
		}

		// Tie up references
		p.account.postings = append(p.account.postings, p)
		p.account.transactions = append(p.account.transactions, t)
	}
}
