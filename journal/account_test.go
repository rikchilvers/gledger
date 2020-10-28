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

	current := root.findOrCreateAccount(components)

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
	savings := root.findOrCreateAccount(components)

	if len(savings.Parent.Children) != 2 {
		t.Fatalf("created account's parent does not have enough children")
	}

	// Search for an account
	searchResult := root.findOrCreateAccount(components)

	if searchResult.Name != "savings" {
		t.Fatalf("search returns incorrect account")
	}

}

func TestAccountPathGeneration(t *testing.T) {
	components := []string{"assets", "current"}
	account := newAccountWithChildren(components, nil)

	if account.path() != "assets:current" {
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
