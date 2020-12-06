package journal

import (
	"fmt"
	"strings"
)

const (
	RootID       string = "_root_"
	BudgetRootID string = "_budget_root_"
)

type Journal struct {
	transactions         []*Transaction
	periodicTransactions []*PeriodicTransaction
	filePaths            []string // the
	root                 *Account
	budgetRoot           *Account
}

func NewJournal() Journal {
	return Journal{
		transactions:         make([]*Transaction, 0, 256),
		periodicTransactions: make([]*PeriodicTransaction, 0, 256),
		filePaths:            make([]string, 0, 10),
		root:                 NewAccount(RootID),
		budgetRoot:           NewAccount(BudgetRootID),
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
			p.Account = j.root.FindOrCreateAccount(pathComponents)
			p.Account.Path = p.AccountPath
			p.Account.PathComponents = pathComponents
		}

		// Add postings to accounts
		p.Account.Postings = append(p.Account.Postings, p)

		// Add the transaction to the account and the journal
		p.Account.Transactions = append(p.Account.Transactions, transaction)
		j.transactions = append(j.transactions, transaction)

		// Add the posting's amount to the account and all its ancestors
		if err := p.Account.WalkAncestors(func(a *Account) error {
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

func (j *Journal) Prepare(showZero bool) {
	if !showZero {
		matcher := func(a Account) bool {
			return a.Amount.Quantity == 0
		}
		matching := j.root.FindAccounts(matcher)
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

func (j *Journal) Report(flattenTree bool) {
	prepender := func(a Account) string {
		return fmt.Sprintf("%20s  ", a.Amount.DisplayableQuantity(true))
	}

	if flattenTree {
		flattened := j.root.FlattenedTree(prepender)
		fmt.Println(flattened)
	} else {
		tree := j.root.Tree(prepender)
		fmt.Println(tree)
	}

	// 20x '-' because that is how wide we format the amount to be
	fmt.Println("--------------------")

	// Print the root account's value
	fmt.Printf("%20s\n", j.root.Amount.DisplayableQuantity(false))
}
