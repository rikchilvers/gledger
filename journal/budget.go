package journal

import (
	"strings"
	"time"
)

// Budget is a wrapper around accounts to enable monthly tracking
type Budget struct {
	Months map[time.Time]BudgetMonth // what was budgeted
}

func NewBudget() Budget {
	return Budget{
		Months: make(map[time.Time]BudgetMonth, 12),
	}
}

type BudgetMonth struct {
	EnvelopeRoot *Account
	ExpenseRoot  *Account
	Income       *Account
	overspending Amount // flows from previous month
	future       Amount // sum of accounts in future months' buckets
}

func newBudgetMonth() BudgetMonth {
	income := NewAccount(IncomeID)
	income.Name = "Funds"
	expenses := NewAccount(ExpensesID)
	expenses.Name = "Spending"
	return BudgetMonth{
		EnvelopeRoot: NewAccount(BudgetRootID),
		ExpenseRoot:  expenses,
		Income:       income,
	}
}

type PostingType int

const (
	EnvelopePosting PostingType = iota
	ExpensePosting
	IncomePosting
)

func (b *Budget) AddPosting(p *Posting, pt PostingType) error {
	if pt == IncomePosting {
		return b.addIncomePosting(p)
	}

	// Get the month
	bm, found := b.Months[normaliseToMonth(p.Transaction.Date)]
	if !found {
		bm = newBudgetMonth()
	}

	defer func() {
		b.Months[normaliseToMonth(p.Transaction.Date)] = bm
	}()

	// wireUpPosting
	if p.Account != nil {
		panic("p.Account != nil")
	}

	// Don't add the BudgetRoot to the budget
	// Instead, roll it in to the EnvelopeRoot
	if pt == EnvelopePosting && p.AccountPath == BudgetRootID {
		bm.EnvelopeRoot.Amount.Add(*p.Amount)
		return nil
	}

	// Choose root
	var root *Account
	switch pt {
	case ExpensePosting:
		root = bm.ExpenseRoot
	case EnvelopePosting:
		root = bm.EnvelopeRoot
	}

	pathComponents := strings.Split(p.AccountPath, ":")
	if pt == ExpensePosting {
		// strip 'Expenses' from the path components
		pathComponents = strings.Split(p.AccountPath, ":")[1:]
	}

	// Assign an account to the posting
	p.Account = root.FindOrCreateAccount(pathComponents)

	// Set commodities
	if bm.EnvelopeRoot.Amount.Commodity == "" {
		bm.EnvelopeRoot.Amount.Commodity = p.Amount.Commodity
	}
	if bm.ExpenseRoot.Amount.Commodity == "" {
		bm.ExpenseRoot.Amount.Commodity = p.Amount.Commodity
	}

	// Make the envelope account
	envelopeAccount := bm.EnvelopeRoot.FindOrCreateAccount(pathComponents)
	envelopeAccount.Amount.Commodity = p.Amount.Commodity

	// Make the expense account
	expenseAccount := bm.ExpenseRoot.FindOrCreateAccount(pathComponents)
	expenseAccount.Amount.Commodity = p.Amount.Commodity

	if pt == EnvelopePosting {
		// Add to the envelope account
		envelopeAccount.Postings = append(envelopeAccount.Postings, p)

		// Add the posting's amount to the envelope account and all of its ancestors
		if err := envelopeAccount.WalkAncestors(func(a *Account) error {
			if err := a.Amount.Add(*p.Amount); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if pt == ExpensePosting {
		// Add to the expense account
		expenseAccount.Postings = append(expenseAccount.Postings, p)

		// As this account is not the same as the non-budget expenses account version
		// we need to ask it to create its path as it drops the 'Expenses:' head
		// TODO might not be necessary? for ExpensePostings?
		if len(envelopeAccount.Path) == 0 {
			envelopeAccount.Path = envelopeAccount.CreatePath()
			envelopeAccount.PathComponents = pathComponents
		}
		if len(expenseAccount.Path) == 0 {
			expenseAccount.Path = expenseAccount.CreatePath()
			expenseAccount.PathComponents = pathComponents
		}

		// Subtract the postings amount from the expense account and all of its ancestors
		if err := expenseAccount.WalkAncestors(func(a *Account) error {
			if err := a.Amount.Subtract(*p.Amount); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (b *Budget) addEnvelopePosting(p *Posting) error {
	return b.AddPosting(p, EnvelopePosting)
}

func (b *Budget) addExpensePosting(p *Posting) error {
	return b.AddPosting(p, ExpensePosting)
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
	if err := bm.Income.Amount.Subtract(*p.Amount); err != nil {
		return err
	}

	if err := bm.EnvelopeRoot.Amount.Subtract(*p.Amount); err != nil {
		return err
	}

	b.Months[normaliseToMonth(p.Transaction.Date)] = bm
	return nil
}

func normaliseToMonth(date time.Time) time.Time {
	return date.AddDate(0, 0, -(date.Day() - 1))
}
