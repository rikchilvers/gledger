// Package cmd handles the cli
package cmd

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/rikchilvers/gledger/journal"
	"github.com/rikchilvers/gledger/parser"
	"github.com/rikchilvers/gledger/reporting"
)

func parse(th parser.TransactionHandler, ph parser.PeriodicTransactionHandler) error {
	if rootJournalPath == "" {
		// TODO: use viper to read env variable
		return errors.New("no root journal path provided")
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

// linkTransaction builds the account tree
func linkTransaction(root *journal.Account, transaction *journal.Transaction, _ string) error {
	for _, p := range transaction.Postings {
		// Use the parsed account path to find or create the account
		if p.Account == nil {
			p.Account = root.FindOrCreateAccount(strings.Split(p.AccountPath, ":"))
		}

		// Add postings to accounts
		p.Account.Postings = append(p.Account.Postings, p)

		// Add the transaction to the account
		p.Account.Transactions = append(p.Account.Transactions, transaction)

		// Add the posting's amount to the account and all its ancestors
		if err := p.Account.WalkAncestors(func(a *journal.Account) error {
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

func dateCheckedFilteringTransactionHandler(args []string, handler func(t *journal.Transaction, path string) error) func(t *journal.Transaction, path string) error {
	return func(t *journal.Transaction, path string) error {
		withinRange, err := withinDateRange(t)
		if err != nil {
			return err
		}

		if !withinRange {
			return nil
		}

		matches, err := matchesRegex(t, args)
		if err != nil {
			return nil
		}

		if !matches {
			return nil
		}

		return handler(t, path)
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

func matchesRegex(t *journal.Transaction, args []string) (bool, error) {
	if len(args) == 0 {
		return true, nil
	}

	for _, arg := range args {
		if !containsUppercase(arg) {
			arg = "(?i)" + arg
		}
		regex, err := regexp.Compile(arg)
		if err != nil {
			return false, err
		}

		// Check payees
		if filterContext.checkPayees {
			if regex.MatchString(t.Payee) {
				return true, nil
			}
		}

		// Check transaction notes
		if filterContext.checkNotes {
			if regex.MatchString(t.HeaderNote) {
				return true, nil
			}
			for _, n := range t.Notes {
				if regex.MatchString(n) {
					return true, nil
				}
			}
		}

		// Check postings
		for _, p := range t.Postings {
			// Check accounts
			if filterContext.checkAccounts {
				if regex.MatchString(p.AccountPath) {
					return true, nil
				}
			}
			// Check postings comments
			if filterContext.checkNotes {
				for _, c := range p.Comments {
					if regex.MatchString(c) {
						return true, nil
					}
				}
			}
		}
	}

	return false, nil
}

func containsUppercase(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return true
		}
	}
	return false
}
