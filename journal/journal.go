package journal

import (
	"strings"
	"time"
)

// Identifiers for accounts
const (
	RootID       string = "_root_"
	BudgetRootID string = "_budget_root_"
	// TODO allow these to be set by the user
	ExpensesID string = "Expenses"
	IncomeID   string = "Income"
)

// Journal holds information about the transactions parsed
type Journal struct {
	config               ProcessingConfig
	transactions         []*Transaction
	periodicTransactions []*PeriodicTransaction
	filePaths            []string
	Root                 *Account
	Budget               Budget
}

// ProcessingConfig contains flags used when parsing transactions
type ProcessingConfig struct {
	CalculateBudget bool
}

// NewJournal creates a Journal
func NewJournal(config ProcessingConfig) Journal {
	j := Journal{
		config:               config,
		transactions:         make([]*Transaction, 0, 256),
		periodicTransactions: make([]*PeriodicTransaction, 0, 256),
		filePaths:            make([]string, 0, 10),
		Root:                 NewAccount(RootID),
		Budget:               NewBudget(),
	}

	return j
}

// AddTransaction adds a transaction to the journal
func (j *Journal) AddTransaction(t *Transaction, locationHint string) {
	j.transactions = append(j.transactions, t)
	j.filePaths = append(j.filePaths, locationHint)
}

// AddPeriodicTransaction adds a periodic transaction to the journal
func (j *Journal) AddPeriodicTransaction(pt *PeriodicTransaction, locationHint string) error {
	j.periodicTransactions = append(j.periodicTransactions, pt)

	if pt.Period.Interval == PNone {
		if j.config.CalculateBudget {
			// Handle budget allocations differently
			// PeriodicTransaction with no interval are considered budget transactions
			// running the pt is necessary to ensure it's transaction has a date
			ts := pt.Run(time.Time{}, time.Time{})[0]
			for _, p := range ts.Postings {
				// if err := wireUpPosting(j.BudgetRoot, transaction, p); err != nil {
				if err := j.Budget.addEnvelopePosting(p); err != nil {
					return err
				}
			}
		}
		return nil
	}

	// Convert the periodic transaction to real transactions then link them
	// TODO take time bounds for running periodic transactions
	transactions := pt.Run(time.Time{}, time.Time{})
	for _, t := range transactions {
		j.AddTransaction(&t, locationHint)

		for _, p := range t.Postings {
			if err := j.AddPosting(p); err != nil {
				return err
			}
		}
	}

	return nil
}

// AddPosting handles adding normal transaction postings to the journal
func (j *Journal) AddPosting(p *Posting) error {
	if err := wireUpPosting(j.Root, p.Transaction, p); err != nil {
		return err
	}

	if j.config.CalculateBudget {
		// Handle budget posting if this posting is an Expense
		if p.Account.PathComponents[0] == ExpensesID {
			if err := j.Budget.addExpensePosting(p); err != nil {
				return err
			}
		}

		// Handle income for budgeting
		if p.Account.PathComponents[0] == IncomeID {
			if err := j.Budget.addIncomePosting(p); err != nil {
				return err
			}
		}
	}

	return nil
}

func wireUpPosting(root *Account, transaction *Transaction, p *Posting) error {
	if p.Account == nil {
		pathComponents := strings.Split(p.AccountPath, ":")
		p.Account = root.FindOrCreateAccount(pathComponents)
	}

	// Add the posting to the account
	p.Account.Postings = append(p.Account.Postings, p)

	// Add the transaction to the account
	p.Account.Transactions = append(p.Account.Transactions, transaction)

	// Add the posting's amount to the account and all of its ancestors
	add := func(a *Account) error {
		// NB: setting the commodity like this will not work with multiple currencies
		if a.Amount.Commodity == "" {
			a.Amount.Commodity = p.Amount.Commodity
		}
		if err := a.Amount.Add(*p.Amount); err != nil {
			return err
		}
		return nil
	}
	if err := p.Account.WalkAncestors(add); err != nil {
		return err
	}

	return nil
}
