package reporting

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
)

type treeContext struct {
	prepender                  func(a journal.Account) string
	depth                      int
	isOnlyChild                bool
	collapsedAccounts          string
	collapsedAccountsDepth     int
	tree                       string
	shouldCollapseOnlyChildren bool
}

func newTreeContext(p func(a journal.Account) string) treeContext {
	return treeContext{
		prepender: p,
	}
}

// Tree walks the descendents of this Account
// and returns a string of its structure in tree form
func Tree(a journal.Account, prepender func(a journal.Account) string) string {
	if prepender == nil {
		prepender = func(a journal.Account) string { return "" }
	}
	c := newTreeContext(prepender)

	for _, childName := range a.SortedChildNames() {
		c = tree(*a.Children[childName], c)
	}

	return c.tree
}

// We keep track of the current line for collapsed only children
// TODO: return two strings (tree, collapsedAccountsLine)
func tree(a journal.Account, c treeContext) treeContext {
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
	if c.isOnlyChild && c.shouldCollapseOnlyChildren {
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
			c = tree(*a.Children[childName], c)
		}
	} else {
		// At this point, we know this account has siblings
		if len(a.Children) == 1 && c.shouldCollapseOnlyChildren {
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
			c = tree(*a.Children[childName], c)
		}
		// Return to previous depth
		c.depth--
	}
	return c
}
