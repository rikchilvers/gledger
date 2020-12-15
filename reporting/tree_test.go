package reporting

import (
	"testing"

	"github.com/rikchilvers/gledger/journal"
)

func TestTree(t *testing.T) {
	p := "£123  "
	expected := `£123  A0
£123    A1a
£123    A1b:A2:A3
£123  E0
£123    E1a:E2a
£123      E3a
£123      E3b
£123    E1b:E2b
£123  I0:I1`
	prepender := func(a journal.Account) string { return p }
	root, _, _ := createRoot()
	got := root.Tree(prepender)

	if got != expected {
		t.Fatalf("\nExpected:\n%s\nGot:\n%s", expected, got)
	}
}
