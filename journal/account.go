package journal

import (
	"fmt"
	"sort"

	"github.com/rikchilvers/gledger/shared"
)

// Account is the
type Account struct {
	Name           string
	Path           string   // Set by newAccountWithChildren
	PathComponents []string // Set by newAccountWithChildren
	Amount         Amount
	Parent         *Account
	Children       map[string]*Account
	Postings       []*Posting
	Transactions   []*Transaction
}

// NewAccount creates an Account
func NewAccount(name string) *Account {
	return &Account{
		Name:           name,
		Path:           "",
		PathComponents: make([]string, 0, 5),
		Parent:         nil,
		Children:       make(map[string]*Account),
		Postings:       make([]*Posting, 0, 2048),
		Transactions:   make([]*Transaction, 0, 1024),
	}
}

func (a *Account) String() string {
	return a.asString(0)
}

// Includes children
func (a *Account) asString(level int) string {
	// Print the name of this account at the specified level
	s := a.Name
	for i := 0; i < level*shared.TabWidth; i++ {
		s = fmt.Sprintf(" %s", s)
	}

	// Print the name of this account's children at the next level
	for _, c := range a.Children {
		s = fmt.Sprintf("%s\n%s", s, c.asString(level+1))
	}

	return s
}

// CreatePath creates a : delimited string from the Account's ancestry
func (a *Account) CreatePath() string {
	path := a.Name
	current := a
	for {
		if current.Parent == nil || current.Parent.Name == RootID || current.Parent.Name == BudgetRootID {
			return path
		}
		path = fmt.Sprintf("%s:%s", current.Parent.Name, path)
		current = current.Parent
	}
}

// Head returns the oldest ancestor that is not root
func (a *Account) Head() *Account {
	current := a
	for {
		if current.Parent == nil || current.Parent.Name == RootID || current.Parent.Name == BudgetRootID {
			return current
		}
		current = current.Parent
	}
}

// WalkAncestors calls `action` on this account and all its ancestors
func (a *Account) WalkAncestors(action func(*Account) error) error {
	if err := action(a); err != nil {
		return err
	}
	if a.Parent == nil {
		return nil
	}
	return a.Parent.WalkAncestors(action)
}

func newAccountWithChildren(parent *Account, components []string) *Account {
	if len(components) == 0 {
		return parent
	}

	child := NewAccount(components[0])
	if parent != nil {
		child.Parent = parent
		parent.Children[child.Name] = child
	}

	if parent == nil || parent.Name == RootID || parent.Name == BudgetRootID {
		child.PathComponents = []string{components[0]}
	} else {
		child.PathComponents = append(parent.PathComponents, components[0])
	}

	child.Path = child.CreatePath()

	return newAccountWithChildren(child, components[1:])
}

// FindOrCreateAccount searches the Account's children for one matching the components,
// creating children as necessary if it does not find matching ones
func (a *Account) FindOrCreateAccount(components []string) *Account {
	child, remainingComponents := a.findChild(components)

	// If there were no remaining accounts, we found the deepest
	if remainingComponents == nil {
		return child
	}

	// Otherwise, add a child to the root account
	// fmt.Println("new account", remainingComponents, child.Name)
	return newAccountWithChildren(child, remainingComponents)
}

// Returns the deepest account found and any remaining components
func (a *Account) findChild(components []string) (*Account, []string) {
	if account, didFind := a.Children[components[0]]; didFind {
		if len(components) > 1 {
			return account.findChild(components[1:])
		}
		// There are no remaining components so we found the deepest child
		return account, nil
	}
	// We didn't find a child and there are remaining components
	// so return those and the deepest parent
	return a, components
}

// SortedChildNames returns an alphabetically sorted slice of the Account's children's names
func (a *Account) SortedChildNames() []string {
	names := make([]string, 0, len(a.Children))
	for name := range a.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// PruneChildren removes child nodes beneath a certain depth
func (a *Account) PruneChildren(targetDepth, currentDepth int) {
	// If we've reached the target depth, remove all children
	if currentDepth == targetDepth {
		for key := range a.Children {
			delete(a.Children, key)
		}
	}
	for _, child := range a.Children {
		child.PruneChildren(targetDepth, currentDepth+1)
	}
}

// Leaves finds all accounts with no children
func (a *Account) Leaves() []*Account {
	matcher := func(a Account) bool {
		return len(a.Children) == 0
	}
	return a.FindAccounts(matcher)
}

// FindAccounts walks the account tree and returns matching accounts
func (a *Account) FindAccounts(matcher func(a Account) bool) []*Account {
	found := make([]*Account, 0, 5)
	found = append(found, a.findAccounts(matcher, found)...)
	return found
}

func (a *Account) findAccounts(matcher func(a Account) bool, found []*Account) []*Account {
	if matcher(*a) {
		found = append(found, a)
	}

	// Descend to this Account's children
	for _, child := range a.Children {
		found = child.findAccounts(matcher, found)
	}

	return found
}

func (a *Account) RemoveEmptyChildren() {
	matcher := func(a Account) bool {
		return a.Amount.Quantity == 0
	}
	matching := a.FindAccounts(matcher)
	for _, m := range matching {
		if m.Name == RootID {
			continue
		}
		m.Unlink()
	}
}

// Unlink removes this account from it's parents
func (a *Account) Unlink() {
	if a.Parent != nil {
		delete(a.Parent.Children, a.Name)
	}
	a.Parent = nil
}

// RemoveChildren removes children which do not return true from the matcher func
func (a *Account) RemoveChildren(matcher func(a Account) bool) bool {
	// Start with whether this account matches
	matches := matcher(*a)

	toUnlink := make([]*Account, 0, len(a.Children))
	for _, child := range a.Children {
		childMatches := child.RemoveChildren(matcher)

		if !childMatches {
			toUnlink = append(toUnlink, child)
		}

		matches = matches || childMatches
	}

	// Remove unmatching children
	for _, child := range toUnlink {
		child.Unlink()
	}

	return matches
}
