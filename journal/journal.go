package journal

import (
	"fmt"
	"strings"
	"time"
)

const (
	RootID         string = "_root_"
	BudgetRootID   string = "_budget_root_"
	ToBeBudgetedID string = "To Be Budgeted"
	// TODO allow these to be set by the user
	ExpensesID string = "Expenses"
	IncomeID   string = "Income"
)

type Journal struct {
	transactions         []*Transaction
	periodicTransactions []*PeriodicTransaction
	filePaths            []string // the
	Root                 *Account
	BudgetRoot           *Account
}

func NewJournal() Journal {
	j := Journal{
		transactions:         make([]*Transaction, 0, 256),
		periodicTransactions: make([]*PeriodicTransaction, 0, 256),
		filePaths:            make([]string, 0, 10),
		Root:                 NewAccount(RootID),
		BudgetRoot:           NewAccount(BudgetRootID),
	}

	tbb := NewAccount(ToBeBudgetedID)
	tbb.Parent = j.BudgetRoot
	j.BudgetRoot.Children[ToBeBudgetedID] = tbb

	return j
}

func (j *Journal) AddTransaction(t *Transaction, locationHint string) error {
	// TODO: make filePaths an indexed map
	return j.linkTransaction(t)
}

func (j *Journal) AddPeriodicTransaction(pt *PeriodicTransaction, locationHint string) error {
	// Add the periodic transaction to the journal
	j.periodicTransactions = append(j.periodicTransactions, pt)

	if pt.Period.Interval == PNone {
		// Handle budget allocations differently
		// PeriodicTransaction with no interval are considered budget transactions
		// TODO gate behind budget flag
		return j.linkBudgetTransaction(&pt.Transaction)
	}

	// Convert the periodic transaction to real transactions then link them
	// TODO take time bounds for running periodic transactions
	transactions := pt.Run(time.Time{}, time.Time{})
	for _, p := range transactions {
		if err := j.linkTransaction(&p); err != nil {
			return err
		}
	}

	return nil
}

func (j *Journal) linkBudgetTransaction(t *Transaction) error {
	fmt.Println("would handle budget transaction")
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

		// Handle budget posting if this posting is an Expense
		// TODO gate behind budget flag
		if p.Account.PathComponents[0] == ExpensesID {
			if err := j.handleExpensesPosting(p); err != nil {
				return err
			}
		}

		// Handle income for budgeting
		// TODO gate behind budget flag
		if p.Account.PathComponents[0] == IncomeID {
			if err := j.handleIncomePosting(p); err != nil {
				return err
			}
		}
	}

	return nil
}

func (j *Journal) handleIncomePosting(posting *Posting) error {
	// Add the income to 'To Be Budgeted'
	tbb := j.BudgetRoot.Children[ToBeBudgetedID]

	// We subtract to make the income positive
	if err := tbb.WalkAncestors(func(a *Account) error {
		if err := a.Amount.Subtract(posting.Amount); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (j *Journal) handleExpensesPosting(posting *Posting) error {
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
