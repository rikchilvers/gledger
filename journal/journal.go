package journal

import (
	"strings"
)

const (
	RootID       string = "_root_"
	BudgetRootID string = "_budget_root_"
	// TODO allow this to be set by the user
	ExpensesID string = "Expenses"
)

type Journal struct {
	transactions         []*Transaction
	periodicTransactions []*PeriodicTransaction
	filePaths            []string // the
	Root                 *Account
	BudgetRoot           *Account
}

func NewJournal() Journal {
	return Journal{
		transactions:         make([]*Transaction, 0, 256),
		periodicTransactions: make([]*PeriodicTransaction, 0, 256),
		filePaths:            make([]string, 0, 10),
		Root:                 NewAccount(RootID),
		BudgetRoot:           NewAccount(BudgetRootID),
	}
}

func (j *Journal) AddTransaction(t *Transaction, locationHint string) error {
	// TODO: make filePaths an indexed map
	return j.linkTransaction(t)
}

func (j *Journal) AddPeriodicTransaction(t *PeriodicTransaction, locationHint string) error {
	return nil
}

func (j *Journal) linkTransaction(transaction *Transaction) error {
	for _, p := range transaction.Postings {
		// Use the parsed account path to find or create the account
		if p.Account == nil {
			pathComponents := strings.Split(p.AccountPath, ":")
			p.Account = j.Root.FindOrCreateAccount(pathComponents)
			p.Account.Path = p.AccountPath
			p.Account.PathComponents = pathComponents
		}

		// Add the posting to its account's postings
		p.Account.Postings = append(p.Account.Postings, p)

		// Add the transaction to the account and the journal
		p.Account.Transactions = append(p.Account.Transactions, transaction)
		j.transactions = append(j.transactions, transaction)

		// Add the posting's amount to the account and all of its ancestors
		if err := p.Account.WalkAncestors(func(a *Account) error {
			if err := a.Amount.Add(p.Amount); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}

		// Handle budget posting
		if p.Account.PathComponents[0] == ExpensesID {
			if err := j.handleBudgetPosting(p); err != nil {
				return err
			}
		}
	}

	return nil
}

func (j *Journal) handleBudgetPosting(posting *Posting) error {
	account := j.BudgetRoot.FindOrCreateAccount(posting.Account.PathComponents)

	// Subtract the posting's amount from the account and all of its ancestors
	if err := account.WalkAncestors(func(a *Account) error {
		if err := a.Amount.Subtract(posting.Amount); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (j *Journal) Prepare(showZero bool) {
	if !showZero {
		matcher := func(a Account) bool {
			return a.Amount.Quantity == 0
		}
		matching := j.Root.FindAccounts(matcher)
		for _, m := range matching {
			if m.Name == RootID {
				continue
			}
			// remove the account from it's parent
			delete(m.Parent.Children, m.Name)
			m.Parent = nil
		}
	}
}
