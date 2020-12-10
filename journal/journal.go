package journal

import (
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
	config               JournalConfig
	transactions         []*Transaction
	periodicTransactions []*PeriodicTransaction
	filePaths            []string // the
	Root                 *Account
	BudgetRoot           *Account
}

type JournalConfig struct {
	CalculateBudget bool
}

func NewJournal(config JournalConfig) Journal {
	j := Journal{
		config:               config,
		transactions:         make([]*Transaction, 0, 256),
		periodicTransactions: make([]*PeriodicTransaction, 0, 256),
		filePaths:            make([]string, 0, 10),
		Root:                 NewAccount(RootID),
		BudgetRoot:           NewAccount(BudgetRootID),
	}

	if config.CalculateBudget {
		tbb := NewAccount(ToBeBudgetedID)
		tbb.Parent = j.BudgetRoot
		j.BudgetRoot.Children[ToBeBudgetedID] = tbb
	}

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
		if j.config.CalculateBudget {
			// Handle budget allocations differently
			// PeriodicTransaction with no interval are considered budget transactions
			return j.linkBudgetTransaction(&pt.Transaction)
		}
		return nil
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

// linkBudgetTransaction wires up allocations to the budget
func (j *Journal) linkBudgetTransaction(transaction *Transaction) error {
	for _, p := range transaction.Postings {
		if err := wireUpPosting(j.BudgetRoot, transaction, p); err != nil {
			return err
		}

		// Subtract the posting's amount from To Be Budgeted
		tbb := j.BudgetRoot.Children[ToBeBudgetedID]
		tbb.Amount.Subtract(p.Amount)
	}

	return nil
}

func (j *Journal) linkTransaction(transaction *Transaction) error {
	for _, p := range transaction.Postings {
		if err := wireUpPosting(j.Root, transaction, p); err != nil {
			return err
		}

		// Add the transaction to the list
		j.transactions = append(j.transactions, transaction)

		if j.config.CalculateBudget {
			// Handle budget posting if this posting is an Expense
			if p.Account.PathComponents[0] == ExpensesID {
				if err := j.handleExpensesPosting(p); err != nil {
					return err
				}
			}

			// Handle income for budgeting
			if p.Account.PathComponents[0] == IncomeID {
				if err := j.handleIncomePosting(p); err != nil {
					return err
				}
			}

		}
	}

	return nil
}

func wireUpPosting(root *Account, transaction *Transaction, p *Posting) error {
	if p.Account == nil {
		pathComponents := strings.Split(p.AccountPath, ":")
		p.Account = root.FindOrCreateAccount(pathComponents)
		p.Account.Path = p.AccountPath
		p.Account.PathComponents = pathComponents
	}

	// Add the posting to its account's postings
	p.Account.Postings = append(p.Account.Postings, p)

	// Add the transaction to the account and the journal
	p.Account.Transactions = append(p.Account.Transactions, transaction)

	// Add the posting's amount to the account and all of its ancestors
	if err := p.Account.WalkAncestors(func(a *Account) error {
		if err := a.Amount.Add(p.Amount); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
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
	pathComponents := posting.Account.PathComponents[1:]
	account := j.BudgetRoot.FindOrCreateAccount(pathComponents)
	// As this account is not the same as the non-budget expenses account version
	// we need to ask it to create its path as it drops the 'Expenses:' head
	account.Path = account.CreatePath()
	account.PathComponents = pathComponents

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
		j.Root.removeEmptyChildren()

		if j.config.CalculateBudget {
			j.BudgetRoot.removeEmptyChildren()
		}
	}
}
