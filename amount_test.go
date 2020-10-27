package main

import "testing"

// Also tests commodities with spaces
func TestDisplaysPositiveAmount(t *testing.T) {
	amount := newAmount(4281)
	amount.commodity = "GBP "

	expected := "GBP 42.81"
	got := amount.displayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "42.81"
	got = amount.displayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

// Also tests commodities without spaces
func TestDisplaysNegativeAmount(t *testing.T) {
	amount := newAmount(-4281)
	amount.commodity = "£"

	expected := "£-42.81"
	got := amount.displayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "-42.81"
	got = amount.displayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

func TestDisplaysThreeDigitAmounts(t *testing.T) {
	amount := newAmount(981)
	amount.commodity = "GBP "

	expected := "GBP 9.81"
	got := amount.displayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "9.81"
	got = amount.displayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

func TestDisplaysTwoDigitAmounts(t *testing.T) {
	amount := newAmount(81)
	amount.commodity = "GBP "

	expected := "GBP 0.81"
	got := amount.displayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "0.81"
	got = amount.displayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

func TestDisplaysOneDigitAmounts(t *testing.T) {
	amount := newAmount(9)
	amount.commodity = "GBP "

	expected := "GBP 0.09"
	got := amount.displayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "0.09"
	got = amount.displayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}
