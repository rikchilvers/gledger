package journal

import (
	"fmt"

	. "github.com/rikchilvers/gledger/shared"
)

type Account struct {
	Name         string
	Amount       *Amount
	Parent       *Account
	Children     map[string]*Account
	Postings     []*Posting
	Transactions []*Transaction
}

func (a Account) String() string {
	return a.asString(0)
}

// Includes children
func (a Account) asString(level int) string {
	// Print the name of this account at the specified level
	s := fmt.Sprintf("%s", a.Name)
	for i := 0; i < level*TabWidth; i++ {
		s = fmt.Sprintf(" %s", s)
	}

	// Print the name of this account's children at the next level
	for _, c := range a.Children {
		s = fmt.Sprintf("%s\n%s", s, c.asString(level+1))
	}

	return s
}

// TODO: set this as a variable from the posting
func (a Account) Path() string {
	path := a.Name
	current := a
	for {
		if current.Parent == nil {
			return path
		}
		path = fmt.Sprintf("%s:%s", current.Parent.Name, path)
		current = *current.Parent
	}
}

func NewAccount(name string) *Account {
	return &Account{
		Name:         name,
		Amount:       NewAmount(0),
		Parent:       nil,
		Children:     make(map[string]*Account),
		Postings:     make([]*Posting, 0, 2048),
		Transactions: make([]*Transaction, 0, 1024),
	}
}

// Adds descending child accounts to a parent
func newAccountWithChildren(components []string, parent *Account) *Account {
	for {
		if len(components) == 0 {
			return parent
		}

		a := NewAccount(components[0])
		if parent != nil {
			a.Parent = parent
			parent.Children[a.Name] = a
		}
		parent = a
		components = components[1:]
	}
}

func (a Account) findOrCreateAccount(components []string) *Account {
	deepest, remaining := a.findChildAndDescend(components)

	// If there were no remaining accounts, we found the deepest
	if remaining == nil {
		return deepest
	}

	// Otherwise, add a child to the root account
	return newAccountWithChildren(remaining, deepest)
}

// Returns the deepest account found and any remaining components
func (a Account) findChildAndDescend(components []string) (*Account, []string) {
	if account, didFind := a.Children[components[0]]; didFind {
		if len(components) > 1 {
			return account.findChildAndDescend(components[1:])
		}
		// There are no remaining components so we found the deepest child
		return account, nil
	}
	// There are remaining components so return those and the deepest parent
	return &a, components
}
