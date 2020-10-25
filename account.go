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
	fmt.Println(a.path())
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
			fmt.Printf("parent of %s is nil\n", path)
			return path
		}
		path = fmt.Sprintf("%s:%s", current.parent.name, path)
		current = *current.parent
	}
}

func (a account) pathComponents() []string {
	components := make([]string, 0, 10)
	current := a
	for {
		if current.parent == nil {
			break
		}
		components = append(components, current.parent.name)
		current = *current.parent
	}
	return reverse(components)
}

// From https://stackoverflow.com/a/61218109
func reverse(s []string) []string {
	a := make([]string, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
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
			parent.addChild(a)
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
	if account, found := a.children[components[0]]; found {
		if len(components) > 1 {
			return account.findChildAndDescend(components[1:])
		}
		// There are no remaining components so we found the deepest child
		return account, nil
	}
	// There are remaining components so return those and the deepest parent
	return &a, components
}

func (a account) findChild(name string) *account {
	if account, found := a.children[name]; found {
		return account
	}
	return nil
}

func (a *account) addChild(account *account) {
	if _, found := a.children[account.name]; found {
		return
	}

	a.children[account.name] = account
}
