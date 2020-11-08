package journal

import (
	"fmt"
	"sort"

	"github.com/rikchilvers/gledger/shared"
)

const RootID string = "_root_"

// Account is the
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
	for i := 0; i < level*shared.TabWidth; i++ {
		s = fmt.Sprintf(" %s", s)
	}

	// Print the name of this account's children at the next level
	for _, c := range a.Children {
		s = fmt.Sprintf("%s\n%s", s, c.asString(level+1))
	}

	return s
}

// Path creates a : delimited string from the Account's ancestry
// TODO: set this as a variable from the posting
func (a Account) Path() string {
	path := a.Name
	current := a
	for {
		if current.Parent == nil || current.Parent.Name == RootID {
			return path
		}
		path = fmt.Sprintf("%s:%s", current.Parent.Name, path)
		current = *current.Parent
	}
}

// NewAccount creates an Account
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

// FindOrCreateAccount searches the Account's children for an one matching the components,
// creating children as necessary if it does not find matching ones
func (a Account) FindOrCreateAccount(components []string) *Account {
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

// SortedChildNames returns an alphabetically sorted slice of the Account's children's names
func (a Account) SortedChildNames() []string {
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

// Tree walks the descendents of this Account
// and returns a string of its structure in tree form
func (a Account) Tree(prepender func(a Account) string) string {
	if prepender == nil {
		prepender = func(a Account) string { return "" }
	}
	c := newTreeContext(prepender)

	for _, childName := range a.SortedChildNames() {
		c = a.Children[childName].tree(c)
	}

	return c.tree
}

type treeContext struct {
	prepender              func(a Account) string
	depth                  int
	isOnlyChild            bool
	collapsedAccounts      string
	collapsedAccountsDepth int
	tree                   string
}

func newTreeContext(p func(a Account) string) treeContext {
	return treeContext{
		prepender: p,
	}
}

// We keep track of the current line for collapsed only children
// TODO: return two strings (tree, collapsedAccountsLine)
// TODO: enable prepender to be nil
func (a Account) tree(c treeContext) treeContext {
	calculateSpaces := func(depth int) string {
		spaces := ""
		var tabWidth int = 2
		for i := 0; i < depth*tabWidth; i++ {
			spaces = fmt.Sprintf(" %s", spaces)
		}
		return spaces
	}

	// Only children are a special case because they are collapsed to a single line
	// For this to work with the prepender, we keep track of it separately from the tree
	// until we reach a leaf, where we can rejoin the tree
	if c.isOnlyChild {
		if len(a.Children) == 1 {
			c.collapsedAccounts = fmt.Sprintf("%s:%s", c.collapsedAccounts, a.Name)
			c.isOnlyChild = true
		} else {
			// If this only child has 0 or >1 children
			// we need to add the line to the tree + the prepended string
			spaces := calculateSpaces(c.collapsedAccountsDepth)
			if len(c.tree) == 0 {
				// If this is the first account, don't add a newline
				c.tree = fmt.Sprintf("%s%s%s:%s", c.prepender(a), spaces, c.collapsedAccounts, a.Name)
			} else {
				c.tree = fmt.Sprintf("%s\n%s%s%s:%s", c.tree, c.prepender(a), spaces, c.collapsedAccounts, a.Name)
			}
			c.collapsedAccounts = ""
			c.isOnlyChild = false
		}

		// Descend to children
		for _, childName := range a.SortedChildNames() {
			c = a.Children[childName].tree(c)
		}
	} else {
		// At this point, we know this account has siblings
		if len(a.Children) == 1 {
			// If we have one child, we should add ourselves to the collapsedAccounts line
			if len(c.collapsedAccounts) == 0 {
				// If we're the first only child, don't add a colon
				c.collapsedAccounts = fmt.Sprintf("%s%s", calculateSpaces(c.depth), a.Name)
			} else {
				c.collapsedAccounts = fmt.Sprintf("%s:%s", c.collapsedAccounts, a.Name)
			}
			c.isOnlyChild = true // let the Account's child know it has no siblings
		} else {
			// If we have 0 or >1 children, we should add ourselves to the tree
			if len(c.tree) == 0 {
				// If this is the first account, don't add a newline
				c.tree = fmt.Sprintf("%s%s%s", c.prepender(a), calculateSpaces(c.depth), a.Name)
			} else {
				c.tree = fmt.Sprintf("%s\n%s%s%s", c.tree, c.prepender(a), calculateSpaces(c.depth), a.Name)
			}
			c.isOnlyChild = false // let the Account's children know they have siblings
		}

		// Descend to children at a depth +1
		c.depth++
		for _, childName := range a.SortedChildNames() {
			c = a.Children[childName].tree(c)
		}
		// Return to previous depth
		c.depth--
	}
	return c
}

// FlattenedTree walks the descendents of this Account
// and returns a string of its structure in flattened tree form
func (a Account) FlattenedTree(prepender func(a Account) string) string {
	if prepender == nil {
		prepender = func(a Account) string { return "" }
	}
	return a.flattenedTree(prepender, "")
}

func (a Account) flattenedTree(prepender func(a Account) string, current string) string {
	// If this account has no children, add its path
	if len(a.Children) == 0 {
		return fmt.Sprintf("%s%s%s\n", current, prepender(a), a.Path())
	}

	// If it does have children, descend to them
	for _, childName := range a.SortedChildNames() {
		current = a.Children[childName].flattenedTree(prepender, current)
	}

	return current
}

func (a Account) Leaves() []*Account {
	matcher := func(a Account) bool {
		return len(a.Children) == 0
	}
	return a.FindAccounts(matcher)
}

func (a Account) FindAccounts(matcher func(a Account) bool) []*Account {
	found := make([]*Account, 0, 5)
	found = append(found, a.findAccounts(matcher, found)...)
	return found
}

func (a Account) findAccounts(matcher func(a Account) bool, found []*Account) []*Account {
	if matcher(a) {
		found = append(found, &a)
	}

	// Descend to this Account's children
	for _, child := range a.Children {
		found = child.findAccounts(matcher, found)
	}

	return found
}
