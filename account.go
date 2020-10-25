package main

import (
	"fmt"
)

type account struct {
	name         string
	amount       *amount
	parent       *account
	children     map[string]*account
	postings     []*posting
	transactions []*transaction
}

func (a account) String() string {
	return a.asString(0)
}

// Includes children
func (a account) asString(level int) string {
	// Print the name of this account at the specified level
	s := fmt.Sprintf("%s", a.name)
	for i := 0; i < level*tabWidth; i++ {
		s = fmt.Sprintf(" %s", s)
	}

	// Print the name of this account's children at the next level
	for _, c := range a.children {
		s = fmt.Sprintf("%s\n%s", s, c.asString(level+1))
	}

	return s
}

func (a account) path() string {
	path := a.name
	current := a
	for {
		if current.parent == nil {
			return path
		}
		path = fmt.Sprintf("%s:%s", current.parent.name, path)
		current = *current.parent
	}
}

func newAccount(name string) *account {
	return &account{
		name:         name,
		amount:       newAmount(0),
		parent:       nil,
		children:     make(map[string]*account),
		postings:     make([]*posting, 0, 2048),
		transactions: make([]*transaction, 0, 1024),
	}
}

// Adds descending child accounts to a parent
func newAccountWithChildren(components []string, parent *account) *account {
	for {
		if len(components) == 0 {
			return parent
		}

		a := newAccount(components[0])
		if parent != nil {
			a.parent = parent
			parent.children[a.name] = a
		}
		parent = a
		components = components[1:]
	}
}

func (a account) findOrCreateAccount(components []string) *account {
	deepest, remaining := a.findChildAndDescend(components)

	// If there were no remaining accounts, we found the deepest
	if remaining == nil {
		return deepest
	}

	// Otherwise, add a child to the root account
	return newAccountWithChildren(remaining, deepest)
}

// Returns the deepest account found and any remaining components
func (a account) findChildAndDescend(components []string) (*account, []string) {
	if account, didFind := a.children[components[0]]; didFind {
		if len(components) > 1 {
			return account.findChildAndDescend(components[1:])
		}
		// There are no remaining components so we found the deepest child
		return account, nil
	}
	// There are remaining components so return those and the deepest parent
	return &a, components
}
