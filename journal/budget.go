package journal

import (
	"time"
)

// Budget is a wrapper around accounts to enable monthly tracking
type Budget struct {
	Months map[time.Time]BudgetMonth // metadata about each month
}

func newBudget() Budget {
	return Budget{
		Months: make(map[time.Time]BudgetMonth, 12),
	}
}

type BudgetMonth struct {
	EnvelopeRoot *Account
	Income       *Account
	overspending Amount // flows from previous month
	future       Amount // sum of accounts in future months' buckets
}

func newBudgetMonth() BudgetMonth {
	income := NewAccount(IncomeID)
	income.Name = "Funds"
	return BudgetMonth{
		EnvelopeRoot: NewAccount(BudgetRootID),
		Income:       income,
	}
}

func (b *Budget) addEnvelopePosting(p *Posting) error {
	bm, found := b.Months[normaliseToMonth(p.Transaction.Date)]
	if !found {
		bm = newBudgetMonth()
	}

	if p.AccountPath == BudgetRootID {
		bm.EnvelopeRoot.Amount.Add(p.Amount)
	} else {
		if err := wireUpPosting(bm.EnvelopeRoot, p.Transaction, p); err != nil {
			return err
		}
	}

	b.Months[normaliseToMonth(p.Transaction.Date)] = bm
	return nil
}

func (b *Budget) addExpensePosting(p *Posting) error {
	bm, found := b.Months[normaliseToMonth(p.Transaction.Date)]
	if !found {
		bm = newBudgetMonth()
	}

	if bm.EnvelopeRoot.Amount.Commodity == "" {
		bm.EnvelopeRoot.Amount.Commodity = p.Amount.Commodity
	}

	// strip 'Expenses' from the path components
	pathComponents := p.Account.PathComponents[1:]
	account := bm.EnvelopeRoot.FindOrCreateAccount(pathComponents)
	account.Postings = append(account.Postings, p)

	// We can use whether the commodity is set to determine if we need to 'initialise' the account
	if account.Amount.Commodity == "" {
		// As this account is not the same as the non-budget expenses account version
		// we need to ask it to create its path as it drops the 'Expenses:' head
		account.Path = account.CreatePath()
		account.PathComponents = pathComponents

		account.Amount.Commodity = p.Amount.Commodity
	}

	// Subtract the postings amount fro the account and all of its ancestors
	if err := account.WalkAncestors(func(a *Account) error {
		if err := a.Amount.Subtract(p.Amount); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	b.Months[normaliseToMonth(p.Transaction.Date)] = bm
	return nil
}

func (b *Budget) addIncomePosting(p *Posting) error {
	bm, found := b.Months[normaliseToMonth(p.Transaction.Date)]
	if !found {
		bm = newBudgetMonth()
	}

	if bm.Income.Amount.Commodity == "" {
		bm.Income.Amount.Commodity = p.Amount.Commodity
	}
	if bm.EnvelopeRoot.Amount.Commodity == "" {
		bm.EnvelopeRoot.Amount.Commodity = p.Amount.Commodity
	}

	bm.Income.Postings = append(bm.Income.Postings, p)

	// We subtract to make the income positive
	if err := bm.Income.Amount.Subtract(p.Amount); err != nil {
		return err
	}

	if err := bm.EnvelopeRoot.Amount.Subtract(p.Amount); err != nil {
		return err
	}

	b.Months[normaliseToMonth(p.Transaction.Date)] = bm
	return nil
}

func normaliseToMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -(date.Day() - 1))
}
