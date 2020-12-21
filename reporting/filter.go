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

type FilterType int

const (
	AccountNameFilter FilterType = iota
	PayeeFilter
	NoteFilter
)

type Filter struct {
	regex      *regexp.Regexp
	FilterType FilterType
}

func NewFilter(arg string) (Filter, error) {
	filter := Filter{}

	switch []rune(arg)[0] {
	case '@':
		filter.FilterType = PayeeFilter
		arg = arg[1:]
	case '=':
		filter.FilterType = NoteFilter
		arg = arg[1:]
	default:
		filter.FilterType = AccountNameFilter
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

func (f Filter) MatchesTransaction(t journal.Transaction) bool {
	switch f.FilterType {
	case PayeeFilter:
		return f.regex.MatchString(t.Payee)
	case NoteFilter:
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
	case AccountNameFilter:
		for _, p := range t.Postings {
			if f.regex.MatchString(p.AccountPath) {
				return true
			}
		}
		// TODO: match child accounts
	}

	return false
}

func (f Filter) MatchesString(s string) bool {
	return f.regex.MatchString(s)
}

func MatchesRegex(t *journal.Transaction, args []string) (bool, error) {
	if len(args) == 0 {
		return true, nil
	}

	for _, arg := range args {
		filter, err := NewFilter(arg)
		if err != nil {
			return false, err
		}

		if filter.MatchesTransaction(*t) {
			return true, nil
		}
	}

	return false, nil
}
