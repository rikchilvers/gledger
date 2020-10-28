package parser

import (
	"testing"

	. "github.com/rikchilvers/gledger/journal"
)

func TestDateItemParsing(t *testing.T) {
	parser := NewParser(nil)

	// Correctly formed dates

	err := parser.parseItem(dateItem, []rune("2020/10/11"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formatted date")
	}

	err = parser.parseItem(dateItem, []rune("2020-10-11"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formatted date")
	}

	err = parser.parseItem(dateItem, []rune("2020.10.11"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formatted date")
	}

	// Malformed dates

	err = parser.parseItem(dateItem, []rune("20201011"))
	if err == nil {
		t.Fatalf("parser does not fail for malformed dates")
	}

	err = parser.parseItem(dateItem, []rune("2020-10-89"))
	if err == nil {
		t.Fatalf("parser does not fail for malformed dates")
	}

	err = parser.parseItem(dateItem, []rune("2020.10"))
	if err == nil {
		t.Fatalf("parser does not fail for malformed dates")
	}
}

func TestAmountItemParsing(t *testing.T) {
	parser := NewParser(nil)
	parser.currentPosting = NewPosting()

	parser.previousItemType = commodityItem
	err := parser.parseItem(amountItem, []rune("42.81"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formed amount: %s", err)
	}
	if parser.currentPosting.Amount.Quantity != 4281 {
		t.Fatalf("parser incorrectly parsed amount")
	}

	parser.previousItemType = commodityItem
	err = parser.parseItem(amountItem, []rune("g81"))
	if err == nil {
		t.Fatalf("parser returns no error for incorrectly formed amount: %s", err)
	}

	parser.previousItemType = commodityItem
	err = parser.parseItem(amountItem, []rune("+42.81"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formed amount: %s", err)
	}
	if parser.currentPosting.Amount.Quantity != 4281 {
		t.Fatalf("parser incorrectly parsed amount")
	}

	parser.previousItemType = commodityItem
	err = parser.parseItem(amountItem, []rune("-42.81"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formed amount: %s", err)
	}
	if parser.currentPosting.Amount.Quantity != -4281 {
		t.Fatalf("parser incorrectly parsed amount")
	}
}
