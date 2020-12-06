package journal

import "strings"

const (
	RootID       string = "_root_"
	BudgetRootID string = "_budget_root_"
)

type Journal struct {
	transactions         []*Transaction
	periodicTransactions []*PeriodicTransaction
	filePaths            []string // the
	root                 *Account
	budgetRoot           *Account
}

func newJournal() Journal {
	return Journal{
		transactions:         make([]*Transaction, 0, 256),
		periodicTransactions: make([]*PeriodicTransaction, 0, 256),
		filePaths:            make([]string, 0, 10),
		root:                 NewAccount(RootID),
		budgetRoot:           NewAccount(BudgetRootID),
	}
}

func (j *Journal) addTransaction(t *Transaction, locationHint string) error {
	return nil
}

func (j *Journal) addPeriodicTransaction(t *PeriodicTransaction, locationHint string) error {
	return nil
}

func (j *Journal) linkTransaction(transaction *Transaction) error {
	for _, p := range transaction.Postings {
		// Use the parsed account path to find or create the account
		if p.Account == nil {
			pathComponents := strings.Split(p.AccountPath, ":")
			p.Account = j.root.FindOrCreateAccount(pathComponents)
			p.Account.Path = p.AccountPath
			p.Account.PathComponents = pathComponents
		}

		// Add postings to accounts
		p.Account.Postings = append(p.Account.Postings, p)

		// Add the transaction to the account and the journal
		p.Account.Transactions = append(p.Account.Transactions, transaction)
		j.transactions = append(j.transactions, transaction)

		// Add the posting's amount to the account and all its ancestors
		if err := p.Account.WalkAncestors(func(a *Account) error {
			if err := a.Amount.Add(p.Amount); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
