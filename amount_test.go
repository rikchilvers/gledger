package main

import "testing"

func TestDisplaysAmount(t *testing.T) {
	amount := newAmount(4281)
	amount.commodity = "GBP "

	expected := "GBP 42.81"
	got := amount.displayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly - expected: %s, got: %s", expected, got)
	}

	expected = "42.81"
	got = amount.displayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly - expected: %s, got: %s", expected, got)
	}

	amount.quantity = -4281
	amount.commodity = "£"

	expected = "£-42.81"
	got = amount.displayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly - expected: %s, got: %s", expected, got)
	}

	expected = "-42.81"
	got = amount.displayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly - expected: %s, got: %s", expected, got)
	}
}
