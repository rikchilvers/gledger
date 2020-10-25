package main

import (
	"testing"
)

func TestNewAccountWithChildren(t *testing.T) {
	components := []string{"assets", "current"}
	account := newAccountWithChildren(components, nil)

	if account.name != "current" {
		t.Fatalf("account name is incorrect")
	}

	if account.parent == nil {
		t.Fatalf("account does not have a parent")
	}

	if account.parent.name != "assets" {
		t.Fatalf("account has incorrect parent")
	}

	if _, didFind := account.parent.children["current"]; !didFind {
		t.Fatalf("account parent has missing child")
	}
}

func TestFindOrCreate(t *testing.T) {
	root := newAccount("root")
	components := []string{"assets", "current"}

	current := root.findOrCreateAccount(components)

	if current == nil {
		t.Fatalf("account was not created")
	}

	if current.name != "current" {
		t.Fatalf("created account has wrong name")
	}

	if current.parent == nil {
		t.Fatalf("created account has no parent")
	}

	if current.parent.name != "assets" {
		t.Fatalf("created account has incorrect parent")
	}

	if current.parent.parent.name != "root" {
		t.Fatalf("created account has incorrect grandparent")
	}

	// Add a second account (branch at assets)
	components[1] = "savings"
	savings := root.findOrCreateAccount(components)

	if len(savings.parent.children) != 2 {
		t.Fatalf("created account's parent does not have enough children")
	}

	// Search for an account
	searchResult := root.findOrCreateAccount(components)

	if searchResult.name != "savings" {
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
