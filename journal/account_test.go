package journal

import (
	"testing"
)

func createRoot() (*Account, string, string) {
	/* Full
	A0
		A1a
		A1b
			A2
				A3
	E0
		E1a
			E2a
				E3a
				E3b
		E1b
			E2b
	I0
		I1
	*/

	/* Flattened
	A0:A1a
	A0:A1b:A2:A3
	E0:E1a:E2a:E3a
	E0:E1a:E2a:E3b
	E0:E1b:E2b
	I0:I1
	*/
	f := `A0:A1a
A0:A1b:A2:A3
E0:E1a:E2a:E3a
E0:E1a:E2a:E3b
E0:E1b:E2b
I0:I1`

	/* Tree
	A0
		A1a
		A1b:A2:A3
	E0
		E1a:E2a
			E3a
			E3b
		E1b:E2b
	I0:I1
	*/
	t := `A0
	A1a
	A1b:A2:A3
E0
	E1a:E2a
		E3a
		E3b
	E1b:E2b
I0:I1`

	/* Leaves
	A1a, A3
	E3a, E3b
	E2b
	I1
	*/

	root := NewAccount(RootID)

	componentsAa := []string{"A0", "A1a"}
	componentsAb := []string{"A0", "A1b", "A2", "A3"}
	componentsEa := []string{"E0", "E1a", "E2a", "E3a"}
	componentsEb := []string{"E0", "E1a", "E2a", "E3b"}
	componentsEc := []string{"E0", "E1b", "E2b"}
	componentsIa := []string{"I0", "I1"}

	aa := root.FindOrCreateAccount(componentsAa)
	ab := root.FindOrCreateAccount(componentsAb)
	ea := root.FindOrCreateAccount(componentsEa)
	eb := root.FindOrCreateAccount(componentsEb)
	ec := root.FindOrCreateAccount(componentsEc)
	ia := root.FindOrCreateAccount(componentsIa)

	aa.Path = aa.CreatePath()
	ab.Path = ab.CreatePath()
	ea.Path = ea.CreatePath()
	eb.Path = eb.CreatePath()
	ec.Path = ec.CreatePath()
	ia.Path = ia.CreatePath()

	return root, f, t
}

func TestNewAccountWithChildren(t *testing.T) {
	components := []string{"assets", "current"}
	account := newAccountWithChildren(components, nil)

	if account.Name != "current" {
		t.Fatalf("account name is incorrect")
	}

	if account.Parent == nil {
		t.Fatalf("account does not have a parent")
	}

	if account.Parent.Name != "assets" {
		t.Fatalf("account has incorrect parent")
	}

	if _, didFind := account.Parent.Children["current"]; !didFind {
		t.Fatalf("account parent has missing child")
	}
}

func TestFindOrCreate(t *testing.T) {
	root := NewAccount(RootID)
	components := []string{"assets", "current"}

	current := root.FindOrCreateAccount(components)

	if current == nil {
		t.Fatalf("account was not created")
	}

	if current.Name != "current" {
		t.Fatalf("created account has wrong name")
	}

	if current.Parent == nil {
		t.Fatalf("created account has no parent")
	}

	if current.Parent.Name != "assets" {
		t.Fatalf("created account has incorrect parent")
	}

	if current.Parent.Parent.Name != RootID {
		t.Fatalf("created account has incorrect grandparent")
	}

	// Add a second account (branch at assets)
	components[1] = "savings"
	savings := root.FindOrCreateAccount(components)

	if len(savings.Parent.Children) != 2 {
		t.Fatalf("created account's parent does not have enough children")
	}

	// Search for an account
	searchResult := root.FindOrCreateAccount(components)

	if searchResult.Name != "savings" {
		t.Fatalf("search returns incorrect account")
	}

	components = []string{"expenses", "groceries"}
	root.FindOrCreateAccount(components)
	if len(root.Children) != 2 {
		t.Fatalf("root does not have enough children")
	}
}

func TestAccountPathGeneration(t *testing.T) {
	components := []string{"assets", "current"}
	account := newAccountWithChildren(components, nil)

	if account.CreatePath() != "assets:current" {
		t.Fatalf("account generates incorrect path")
	}
}

func TestAccountPrinting(t *testing.T) {
	components := []string{"assets", "my savings account "}
	account := newAccountWithChildren(components, nil)

	result := account.Parent.String()
	expected := "assets\n  my savings account "

	if result != expected {
		t.Fatalf("account printing does not work\n\texpected %s\nbut got %s", expected, result)
	}
}

func TestMatcher(t *testing.T) {
	root, _, _ := createRoot()
	leaves := root.Leaves()
	expected := 6

	if len(leaves) != expected {
		t.Fatalf("Incorrect number of leaves: expected %d, got %d", expected, len(leaves))
	}
}

func TestPruning(t *testing.T) {
	// For depth zero, we shouldn't see any accounts
	root, _, _ := createRoot()
	root.PruneChildren(0, 0)
	expected := 0
	got := len(root.Children)
	if got != expected {
		t.Fatalf("pruning failed: expected %d, got %d", expected, got)
	}

	// For depth one, we should only see A1 and A2
	root, _, _ = createRoot()
	root.PruneChildren(1, 0)
	expected = 3
	got = len(root.Children)
	if got != expected {
		t.Fatalf("pruning failed: expected %d, got %d", expected, got)
	}
	for _, child := range root.Children {
		if len(child.Children) > 0 {
			t.Fatalf("pruning failed")
		}
	}

	// For depth two, we should see all but C1
	root, _, _ = createRoot()
	root.PruneChildren(2, 0)
	expected = 3
	got = len(root.Children)
	// Check A1 and A2
	if got != expected {
		t.Fatalf("pruning failed: expected %d, got %d", expected, got)
	}

	for name, child := range root.Children {
		if name == "A0" || name == "E0" {
			if len(child.Children) != 2 {
				t.Fatalf("pruning failed")
			}
		}

		if name == "I0" {
			if len(child.Children) != 1 {
				t.Fatalf("pruning failed")
			}
		}
	}
}

func TestRemoveChildren(t *testing.T) {
	root := newTestAccount()

	matcher := func(a Account) bool {
		return a.Name == "Groceries"
	}

	root.RemoveChildren(matcher)

	if len(root.Children) != 1 {
		t.Fatalf("did not remove assets and income")
	}

	expenses, ok := root.Children["Expenses"]
	if !ok || len(expenses.Children) != 1 {
		t.Fatalf("did not find expenses or it had too many children")
	}

	life, ok := expenses.Children["Life"]
	if !ok || len(life.Children) != 1 {
		t.Fatalf("did not find life or it had too many children")
	}

	_, ok = life.Children["Groceries"]
	if !ok {
		t.Fatalf("did not find groceries")
	}
}
