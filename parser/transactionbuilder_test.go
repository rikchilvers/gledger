package parser

import (
	"testing"

	"github.com/rikchilvers/gledger/journal"
	// . "github.com/rikchilvers/gledger/journal"
)

func TestDateItemParsing(t *testing.T) {
	builder := newTransactionBuilder()
	builder.beginTransaction(normalTransaction)

	// Correctly formed dates

	err := builder.build(dateItem, []rune("2020/10/11"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formatted date")
	}

	err = builder.build(dateItem, []rune("2020-10-11"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formatted date")
	}

	err = builder.build(dateItem, []rune("2020.10.11"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formatted date")
	}

	// Malformed dates

	err = builder.build(dateItem, []rune("20201011"))
	if err == nil {
		t.Fatalf("parser does not fail for malformed dates")
	}

	err = builder.build(dateItem, []rune("2020-10-89"))
	if err == nil {
		t.Fatalf("parser does not fail for malformed dates")
	}

	err = builder.build(dateItem, []rune("2020.10"))
	if err == nil {
		t.Fatalf("parser does not fail for malformed dates")
	}
}

func TestAmountItemParsing(t *testing.T) {
	builder := newTransactionBuilder()
	builder.beginTransaction(normalTransaction)
	builder.currentPosting = journal.NewPosting()

	// Correctly formed amounts

	builder.previousItemType = commodityItem
	err := builder.build(amountItem, []rune("42.81"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formed amount: %s", err)
	}
	if builder.currentPosting.Amount.Quantity != 4281 {
		t.Fatalf("parser incorrectly parsed amount")
	}

	builder.previousItemType = commodityItem
	err = builder.build(amountItem, []rune("+42.81"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formed amount: %s", err)
	}
	if builder.currentPosting.Amount.Quantity != 4281 {
		t.Fatalf("parser incorrectly parsed amount")
	}

	builder.previousItemType = commodityItem
	err = builder.build(amountItem, []rune("-42"))
	if err != nil {
		t.Fatalf("parser returns error for correctly formed amount: %s", err)
	}
	got := builder.currentPosting.Amount.Quantity
	expected := -42
	if builder.currentPosting.Amount.Quantity != -4200 {
		t.Fatalf("parser incorrectly parsed amount - expected %d got %d", expected, got)
	}

	// Malformed amounts

	builder.previousItemType = commodityItem
	err = builder.build(amountItem, []rune("g81"))
	if err == nil {
		t.Fatalf("parser returns no error for incorrectly formed amount: %s", err)
	}

	builder.previousItemType = commodityItem
	err = builder.build(amountItem, []rune("8g1"))
	if err == nil {
		t.Fatalf("parser returns no error for incorrectly formed amount: %s", err)
	}
}
