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
	EnvelopeRoot   *Account
	incomePostings []*Posting // posting from Income:**:*
	income         Amount     // sum of income postings amounts
	overspending   Amount     // flows from previous month
	future         Amount     // sum of accounts in future months' buckets
}

func newBudgetMonth() BudgetMonth {
	return BudgetMonth{
		EnvelopeRoot:   NewAccount(BudgetRootID),
		incomePostings: make([]*Posting, 0, 10),
	}
}

func (b *Budget) addEnvelopePosting(p *Posting) error {
	month, found := b.Months[normaliseToMonth(p.Transaction.Date)]
	if !found {
		month = newBudgetMonth()
	}

	if p.AccountPath == BudgetRootID {
		month.EnvelopeRoot.Amount.Add(p.Amount)
	} else {
		if err := wireUpPosting(month.EnvelopeRoot, p.Transaction, p); err != nil {
			return err
		}
	}

	b.Months[normaliseToMonth(p.Transaction.Date)] = month
	return nil
}

func (b *Budget) addIncomePosting(p *Posting) error {
	bm, found := b.Months[normaliseToMonth(p.Transaction.Date)]
	if !found {
		bm = newBudgetMonth()
	}

	bm.incomePostings = append(bm.incomePostings, p)
	// We subtract to make the income positive
	if bm.income.Commodity == "" {
		bm.income.Commodity = p.Amount.Commodity
	}
	if err := bm.income.Subtract(p.Amount); err != nil {
		return err
	}

	b.Months[normaliseToMonth(p.Transaction.Date)] = bm
	return nil
}

func (b *Budget) addExpensePosting(p *Posting) error {
	bm, found := b.Months[normaliseToMonth(p.Transaction.Date)]
	if !found {
		bm = newBudgetMonth()
	}

	// strip 'Expenses' from the path components
	pathComponents := p.Account.PathComponents[1:]
	account := bm.EnvelopeRoot.FindOrCreateAccount(pathComponents)

	if bm.EnvelopeRoot.Amount.Commodity == "" {
		bm.EnvelopeRoot.Amount.Commodity = p.Amount.Commodity
	}

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

func normaliseToMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -(date.Day() - 1))
}
