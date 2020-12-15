package parser

import (
	"testing"
	"time"
)

func TestParseYear(t *testing.T) {
	input := "2020"
	expected := time.Date(2020, time.January, 1, 0, 0, 0, 0, time.Local)
	got, err := ParseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}
}

func TestParseYearIncorrectLength(t *testing.T) {
	input := "20"
	_, err := ParseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year input '%s'", input)
	}
}

func TestParseYearMonth(t *testing.T) {
	input := "2020/06"
	expected := time.Date(2020, time.June, 1, 0, 0, 0, 0, time.Local)
	got, err := ParseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year/month:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year/month\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}

	input = "2020-06"
	expected = time.Date(2020, time.June, 1, 0, 0, 0, 0, time.Local)
	got, err = ParseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year/month:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year/month\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}
}

func TestParseYearMonthIncorrectLength(t *testing.T) {
	input := "20/06"
	_, err := ParseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year/month input '%s'", input)
	}
}

func TestParseYearMonthIncorrectFormat(t *testing.T) {
	input := "2020/42"
	_, err := ParseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year/month input '%s'", input)
	}
}

func TestParseMonthDay(t *testing.T) {
	input := "06/22"
	expected := time.Date(2020, time.June, 22, 0, 0, 0, 0, time.Local)
	got, err := ParseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year/month:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year/month\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}

	input = "06.22"
	expected = time.Date(2020, time.June, 22, 0, 0, 0, 0, time.Local)
	got, err = ParseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year/month:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year/month\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}
}

func TestParseMonthDayIncorrectLength(t *testing.T) {
	input := "20/06"
	_, err := ParseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year/month input '%s'", input)
	}
}

func TestParseMonthDayIncorrectForamt(t *testing.T) {
	input := "13/06"
	_, err := ParseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year/month input '%s'", input)
	}
}

func TestParseYearMonthDay(t *testing.T) {
	input := "2020/06/22"
	expected := time.Date(2020, time.June, 22, 0, 0, 0, 0, time.Local)
	got, err := ParseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year/month:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year/month\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}

	input = "2020.06.22"
	expected = time.Date(2020, time.June, 22, 0, 0, 0, 0, time.Local)
	got, err = ParseSmartDate(input)
	if err != nil {
		t.Fatalf("failed to parse year/month:\nerr: %s", err)
	}
	if got != expected {
		t.Fatalf("failed to parse year/month\nexpected\t'%s'\ngot\t\t'%s'", expected, got)
	}
}

func TestParseYearMonthDayIncorrectLength(t *testing.T) {
	input := "2020/6/3"
	_, err := ParseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year/month input '%s'", input)
	}
}

func TestParseYearMonthDayIncorrectForamt(t *testing.T) {
	input := "20/12/06"
	_, err := ParseSmartDate(input)
	if err == nil {
		t.Fatalf("should have errored for year/month input '%s'", input)
	}
}
