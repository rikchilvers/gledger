package main

import (
	"fmt"
	"strings"
)

type journal struct {
	rootAccount  *account
	transactions []*transaction
}

func newJournal() *journal {
	return &journal{
		rootAccount:  newAccount("root"),
		transactions: make([]*transaction, 0, 1024),
	}
}

// Adds descending child accounts to a parent
func (j journal) newAccountWithChildren(components []string, start *account) *account {
	if start == nil {
		start = j.rootAccount
	}

	for {
		if len(components) == 0 {
			return start
		}

		a := newAccount(components[0])
		start.addChild(a)
		start = a
		components = components[1:]
	}
}

func (j journal) findOrCreateAccount(name string) *account {
	components := strings.Split(name, ":")
	deepest, remaining := j.rootAccount.findChildAndDescend(components)

	// If there were no remaining accounts, we found the deepest
	if remaining == nil {
		fmt.Println("found deepest:", deepest.name)
		return deepest
	}

	// Otherwise, add a child to the root account
	return j.newAccountWithChildren(remaining, deepest)
}

func (j *journal) addTransaction(t *transaction) {
	j.transactions = append(j.transactions, t)
}
