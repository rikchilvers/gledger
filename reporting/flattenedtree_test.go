package reporting

import (
	"testing"

	"github.com/rikchilvers/gledger/journal"
)

func TestFlattenedTree(t *testing.T) {
	prepender := func(a journal.Account) string { return "" }
	root, expected, _ := createRoot()
	got := root.FlattenedTree(prepender)

	if got != expected {
		t.Fatalf("\nExpected:\n'%s'\nGot:\n'%s'", expected, got)
	}
}
