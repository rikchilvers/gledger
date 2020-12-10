// Package reporting handles printing information to the terminal
package reporting

import (
	"fmt"

	"github.com/rikchilvers/gledger/journal"
)

// FlattenedTree walks the descendents of this Account
// and returns a string of its structure in flattened tree form
func FlattenedTree(a journal.Account, prepender func(a journal.Account) string) string {
	if prepender == nil {
		prepender = func(a journal.Account) string { return "" }
	}
	return flattenedTree(a, prepender, "")
}

func flattenedTree(a journal.Account, prepender func(a journal.Account) string, current string) string {
	// If this account has no children, add its path
	if len(a.Children) == 0 {
		if len(current) == 0 {
			// Don't add a newline at the start
			return fmt.Sprintf("%s%s%s", current, prepender(a), a.Path)
		} else {
			return fmt.Sprintf("%s\n%s%s", current, prepender(a), a.Path)
		}
	}

	// If it does have children, descend to them
	for _, childName := range a.SortedChildNames() {
		current = flattenedTree(*a.Children[childName], prepender, current)
	}

	return current
}
