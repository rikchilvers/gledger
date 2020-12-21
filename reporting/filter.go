package reporting

import (
	"regexp"
	"unicode"

	"github.com/rikchilvers/gledger/journal"
)

func ContainsUppercase(s string) bool {
	for _, c := range s {
		if unicode.IsUpper(c) {
			return true
		}
	}
	return false
}

type filterType int

const (
	accountNameFilter filterType = iota
	payeeFilter
	noteFilter
)

type filter struct {
	regex      *regexp.Regexp
	filterType filterType
}

func newFilter(arg string) (filter, error) {
	filter := filter{}

	switch []rune(arg)[0] {
	case '@':
		filter.filterType = payeeFilter
	case '=':
		filter.filterType = noteFilter
	default:
		filter.filterType = accountNameFilter
	}

	if !ContainsUppercase(arg) {
		arg = "(?i)" + arg
	}
	regex, err := regexp.Compile(arg)
	if err != nil {
		return filter, err
	}

	filter.regex = regex

	return filter, nil
}

func (f filter) matchesTransaction(t journal.Transaction) bool {
	switch f.filterType {
	case payeeFilter:
		return f.regex.MatchString(t.Payee)
	case noteFilter:
		if f.regex.MatchString(t.HeaderNote) {
			return true
		}
		for _, n := range t.Notes {
			if f.regex.MatchString(n) {
				return true
			}
		}
		for _, p := range t.Postings {
			for _, n := range p.Comments {
				if f.regex.MatchString(n) {
					return true
				}
			}
		}
	case accountNameFilter:
		for _, p := range t.Postings {
			if f.regex.MatchString(p.AccountPath) {
				return true
			}
		}
		// TODO: match child accounts
	}

	return false
}

func (f filter) matchesString(s string) bool {
	return f.regex.MatchString(s)
}

func MatchesRegex(t *journal.Transaction, args []string) (bool, error) {
	if len(args) == 0 {
		return true, nil
	}

	for _, arg := range args {
		filter, err := newFilter(arg)
		if err != nil {
			return false, err
		}

		if filter.matchesTransaction(*t) {
			return true, nil
		}
	}

	return false, nil
}
