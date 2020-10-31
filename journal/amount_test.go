package journal

import "testing"

// Also tests commodities with spaces
func TestDisplaysPositiveAmount(t *testing.T) {
	amount := NewAmount(4281)
	amount.Commodity = "GBP "

	expected := "GBP 42.81"
	got := amount.DisplayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "42.81"
	got = amount.DisplayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

// Also tests commodities without spaces
func TestDisplaysNegativeAmount(t *testing.T) {
	amount := NewAmount(-4281)
	amount.Commodity = "£"

	expected := "£-42.81"
	got := amount.DisplayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "-42.81"
	got = amount.DisplayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

func TestDisplaysThreeDigitAmounts(t *testing.T) {
	amount := NewAmount(981)
	amount.Commodity = "GBP "

	expected := "GBP 9.81"
	got := amount.DisplayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "9.81"
	got = amount.DisplayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

func TestDisplaysTwoDigitAmounts(t *testing.T) {
	amount := NewAmount(81)
	amount.Commodity = "GBP "

	expected := "GBP 0.81"
	got := amount.DisplayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "0.81"
	got = amount.DisplayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}

func TestDisplaysOneDigitAmounts(t *testing.T) {
	amount := NewAmount(9)
	amount.Commodity = "GBP "

	expected := "GBP 0.09"
	got := amount.DisplayableQuantity(true)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}

	expected = "0.09"
	got = amount.DisplayableQuantity(false)
	if got != expected {
		t.Fatalf("amount displays incorrectly: expected: %s, got: %s", expected, got)
	}
}
