package parser

import (
	"testing"
	"time"
)

func TestParseYear(t *testing.T) {
	input := "2020"
	expected := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)
	got, err := parseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}
}

func TestParseYearIncorrectLength(t *testing.T) {
	input := "20"
	_, err := parseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year input '%s'", input)
	}
}
