package journal

import (
	"fmt"
	"testing"
)

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
	root := NewAccount("root")
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

	if current.Parent.Parent.Name != "root" {
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

}

func TestAccountPathGeneration(t *testing.T) {
	components := []string{"assets", "current"}
	account := newAccountWithChildren(components, nil)

	if account.Path() != "assets:current" {
		t.Fatalf("account generates incorrect path")
	}
}

func TestAccountPrinting(t *testing.T) {
	components := []string{"assets", "my savings account "}
	account := newAccountWithChildren(components, nil)

	result := fmt.Sprintf("%s", account.Parent)
	expected := "assets\n  my savings account "

	if result != expected {
		t.Fatalf("account printing does not work\n\texpected %s\nbut got %s", expected, result)
	}
}

func TestPruning(t *testing.T) {
	createRoot := func() *Account {
		componentsA := []string{"A1", "B1", "C1"}
		componentsB := []string{"A1", "B2"}
		componentsC := []string{"A2", "B3"}
		root := NewAccount("root")
		root.FindOrCreateAccount(componentsA)
		root.FindOrCreateAccount(componentsB)
		root.FindOrCreateAccount(componentsC)
		return root
	}

	// For depth zero, we shouldn't see any accounts
	depthZeroRoot := createRoot()
	depthZeroRoot.PruneChildren(0, 0)
	if len(depthZeroRoot.Children) != 0 {
		t.Fatalf("pruning failed")
	}

	// For depth one, we should only see A1 and A2
	depthOneRoot := createRoot()
	depthOneRoot.PruneChildren(1, 0)
	if len(depthOneRoot.Children) != 2 {
		t.Fatalf("pruning failed")
	}
	for _, child := range depthOneRoot.Children {
		if len(child.Children) > 0 {
			t.Fatalf("pruning failed")
		}
	}

	// For depth two, we should see all but C1
	depthTwoRoot := createRoot()
	depthTwoRoot.PruneChildren(2, 0)
	// Check A1 and A2
	if len(depthTwoRoot.Children) != 2 {
		t.Fatalf("pruning failed")
	}

	for name, child := range depthTwoRoot.Children {
		// Check B1 and B2
		if name == "A1" {
			if len(child.Children) != 2 {
				t.Fatalf("pruning failed")
			}
		}

		// Check B3
		if name == "A2" {
			if len(child.Children) != 1 {
				t.Fatalf("pruning failed")
			}
		}

		// Check C1 does not exist
		for _, grandChild := range child.Children {
			if len(grandChild.Children) != 0 {
				t.Fatalf("pruning failed")
			}
		}
	}
}
