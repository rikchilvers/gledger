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

func TestAccountPathGeneration(t *testing.T) {
	components := []string{"assets", "current"}
	account := newAccountWithChildren(components, nil)

	if account.path() != "assets:current" {
		t.Fatalf("account generates incorrect path")
	}
}
