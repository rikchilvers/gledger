// Package journal includes data for journals
package journal

import (
	"fmt"
	"time"
)

// PeriodType describes when a periodic transaction occurs
//go:generate stringer -type=PeriodType
type PeriodType int

// PeriodType describes the interval of a transaction period
const (
	PNone PeriodType = iota
	PDaily
	PWeekly
	PMonthly
	PQuarterly
	PYearly
	PBiweekly
	PFortnightly
	PBiMonthly
)

// Period describes the duration, interval and frequency of a transaction
type Period struct {
	StartDate         time.Time
	EndDate           time.Time
	Interval          PeriodType
	IntervalFrequency int // the N in 'every N days'
}

// PeriodicTransaction wraps a Transaction and a Period
// A BudgetTransaction is a PeriodicTransaction where the end date is set automatically
type PeriodicTransaction struct {
	Period      Period
	Transaction Transaction
}

// NewPeriodicTransaction creates a new PeriodicTransaction
func NewPeriodicTransaction() PeriodicTransaction {
	return PeriodicTransaction{}
}

// Run converts a single PeriodicTransaction into an array of Transactions for a given date span
// Does not extend time bounds to match parameters
func (pt *PeriodicTransaction) Run(start, end time.Time) []Transaction {
	if pt.Period.Interval == PNone {
		pt.Transaction.Date = pt.Period.StartDate
		return []Transaction{pt.Transaction}
	}

	// Check if this transaction is outside the time bounds
	if (!start.IsZero() && start.After(pt.Period.EndDate)) ||
		(!end.IsZero() && end.Before(pt.Period.StartDate)) {
		return []Transaction{}
	}

	// Sync provided bounds with this transaction's ones
	if !start.IsZero() && start.Before(pt.Period.StartDate) {
		start = pt.Period.StartDate
	}
	if !end.IsZero() && end.After(pt.Period.EndDate) {
		end = pt.Period.EndDate
	}

	fmt.Println("unhandled conversion of non-budget periodic transaction")
	return []Transaction{}
}
