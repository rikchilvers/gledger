// Package cmd handles the cli
package cmd

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/rikchilvers/gledger/journal"
	"github.com/rikchilvers/gledger/parser"
	"github.com/rikchilvers/gledger/reporting"
)

func parse(th parser.TransactionHandler, ph parser.PeriodicTransactionHandler) error {
	if len(rootJournalPath) == 0 {
		path, found := os.LookupEnv("LEDGER_FILE")
		if !found {
			return errors.New("no root journal path provided")
		}
		rootJournalPath = path
	}

	file, err := os.Open(rootJournalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	p := parser.NewParser(th, ph)
	if err := p.Parse(file, rootJournalPath); err != nil {
		return err
	}
	return nil
}

// report prints the given account and it's descendents
// TODO: move this to reporting package
func report(account journal.Account, flattenTree, shouldCollapseOnlyChildren bool) {
	prepender := func(a journal.Account) string {
		return fmt.Sprintf("%20s  ", a.Amount.DisplayableQuantity(true))
	}

	if flattenTree {
		flattened := reporting.FlattenedTree(account, prepender)
		fmt.Println(flattened)
	} else {
		tree := reporting.Tree(account, prepender, shouldCollapseOnlyChildren)
		fmt.Println(tree)
	}

	// 20x '-' because that is how wide we format the amount to be
	fmt.Println("--------------------")

	// Print the root account's value
	fmt.Printf("%20s\n", account.Amount.DisplayableQuantity(false))
}

// dateCheckedTransactionHandler wraps a transaction handler in --begin / --end checks
func dateCheckedTransactionHandler(handler func(t *journal.Transaction, path string) error) func(t *journal.Transaction, path string) error {
	return func(t *journal.Transaction, path string) error {
		withinRange, err := withinDateRange(t)
		if err != nil {
			return err
		}

		if withinRange {
			return handler(t, path)
		}

		return nil
	}
}

func dateCheckedFilteringTransactionHandler(handler func(t *journal.Transaction, path string) error) func(t *journal.Transaction, path string) error {
	return func(t *journal.Transaction, path string) error {
		withinRange, err := withinDateRange(t)
		if err != nil {
			return err
		}

		if !withinRange {
			return nil
		}

		if len(filters) == 0 {
			return handler(t, path)
		}

		for _, f := range filters {
			if f.MatchesTransaction(*t) {
				return handler(t, path)
			}
		}

		return nil
	}
}

func withinDateRange(t *journal.Transaction) (bool, error) {
	var err error
	var start, end time.Time

	if len(beginDate) > 0 && !current {
		start, err = parser.ParseSmartDate(beginDate)
		if err != nil {
			return false, err
		}
	}

	if len(endDate) > 0 && !current {
		end, err = parser.ParseSmartDate(endDate)
		if err != nil {
			return false, err
		}
	}

	if current {
		start = time.Time{}
		end = time.Now().AddDate(0, 0, 1)
	}

	withinRange := (t.Date.Equal(start) || t.Date.After(start)) && (end.IsZero() || t.Date.Before(end))

	return withinRange, nil
}

// checkAgainstFilters lets you know if the transaction matched the filters or
// (in the case that transaction wide attributes didn't match)
// which postings matched the filters
func checkAgainstFilters(t *journal.Transaction) (matchedTransaction bool, postings []*journal.Posting, err error) {
	ok, err := withinDateRange(t)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}

	if len(filters) == 0 {
		return true, t.Postings, nil
	}

	matchedPostings := make(map[*journal.Posting]bool)
	for _, f := range filters {
		matchesPayee, matchesTransactionNote, mp := f.MatchesTransactionHow(*t)

		if matchesPayee || matchesTransactionNote {
			// If we've matched on something transaction-wide, we can return everything here
			return true, t.Postings, nil
		}

		for _, p := range mp {
			matchedPostings[p] = true
		}
	}
	for k := range matchedPostings {
		postings = append(postings, k)
	}

	return false, postings, nil
}
