package parser

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/rikchilvers/gledger/journal"
)

// TransactionHandler is the func commands can use to analyse the journal.
// Takes a Transaction and a path to the file where this Transaction was found.
type (
	TransactionHandler         = func(t *journal.Transaction, path string) error
	PeriodicTransactionHandler = func(t *journal.PeriodicTransaction, path string) error
	itemParser                 = func(t itemType, content []rune) error
)

// Parser is how gledger reads journal files
type Parser struct {
	transactionHandler         TransactionHandler
	periodicTransactionHandler PeriodicTransactionHandler
	transactionBuilder         transactionBuilder
	journalFiles               []string
}

// NewParser creates a parser (including its journal)
func NewParser(th TransactionHandler, ph PeriodicTransactionHandler) Parser {
	return Parser{
		transactionHandler:         th,
		periodicTransactionHandler: ph,
		transactionBuilder:         newTransactionBuilder(),
		journalFiles:               make([]string, 0, 2),
	}
}

// Parse lexes and parses the provided file line by line
func (p *Parser) Parse(reader io.Reader, locationHint string) error {
	p.journalFiles = append(p.journalFiles, locationHint)

	// Begin lexing
	lexer := newLexer(reader, locationHint, p.parseItem)
	if err := lexer.lex(); err != nil {
		// This is the exit point for the lexer's errors
		return fmt.Errorf("error at %w", err)
	}

	return nil
}

func (p *Parser) parseItem(t itemType, content []rune) error {
	switch t {
	case emptyLineItem:
		// an empty line signals that the transaction should close
		if err := p.transactionBuilder.endTransaction(*p); err != nil {
			return err
		}
	case eofItem: // Make sure we close the final transaction
		if err := p.transactionBuilder.endTransaction(*p); err != nil {
			return err
		}
	case includeItem:
		path := string(content)

		// Add the file to the slice
		p.journalFiles = append(p.journalFiles, path)

		// Open the file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Lex the file
		lexer := newLexer(file, path, p.parseItem)
		if err := lexer.lex(); err != nil {
			// This is the exit point for the lexer's errors
			return fmt.Errorf("error at %w", err)
		}

		// Drop the file from the slice
		p.journalFiles = p.journalFiles[:1]
	case dateItem:
		// This will start a transaction so check if we need to close a previous one
		// in case there is no empty line between transactions
		if err := p.transactionBuilder.endTransaction(*p); err != nil {
			return err
		}

		p.transactionBuilder.beginTransaction(normalTransaction)

		// tell the builder about this date
		if err := p.transactionBuilder.build(t, content); err != nil {
			return fmt.Errorf("error parsing date\n%w", err)
		}
	case periodItem:
		// This will start a transaction so check if we need to close a previous one
		// in case there is no empty line between transactions
		if err := p.transactionBuilder.endTransaction(*p); err != nil {
			return err
		}

		p.transactionBuilder.beginTransaction(periodicTransaction)
		p.transactionBuilder.build(t, content)
	default:
		return p.transactionBuilder.build(t, content)
	}

	return nil
}

func parseDate(content []rune) (time.Time, error) {
	const dashDateFormat string = "2006-01-02"
	const dotdateItemFormat string = "2006.01.02"
	const SlashDateFormat string = "2006/01/02"

	s := string(content)
	var date time.Time
	var err error

	switch content[4] {
	case '-':
		date, err = time.Parse(dashDateFormat, s)
	case '.':
		date, err = time.Parse(dotdateItemFormat, s)
	case '/':
		date, err = time.Parse(SlashDateFormat, s)
	default:
		return time.Time{}, fmt.Errorf("date is malformed: %s", s)
	}

	if err != nil {
		return time.Time{}, err
	}

	return date, nil
}

func parseAmount(content []rune) (int64, error) {
	// Handle signs
	firstRune := content[0]
	var multiplier int64
	if firstRune == '+' {
		multiplier = 1
		content = content[1:]
	} else if firstRune == '-' {
		multiplier = -1
		content = content[1:]
	} else {
		multiplier = 1
	}

	// Find out if we have a decimal number
	decimalPosition := -1
	for i, r := range content {
		if r == '.' {
			decimalPosition = i
			break
		}
	}

	// Find whole and decimal
	var whole int64
	var decimal int64
	var err error
	if decimalPosition == -1 {
		// TODO: consider https://stackoverflow.com/a/29255836
		whole, err = strconv.ParseInt(string(content), 10, 64)
		decimal = 0

		if err != nil {
			return 0, err
		}
	} else {
		whole, err = strconv.ParseInt(string(content[:decimalPosition]), 10, 64)
		if err != nil {
			return 0, err
		}
		decimal, err = strconv.ParseInt(string(content[decimalPosition+1:]), 10, 64)
		if err != nil {
			return 0, err
		}
	}

	quantity := multiplier * (whole*100 + decimal)

	return quantity, nil
}

func parsePeriod(content []rune) (journal.Period, error) {
	const budgetDateFormat string = "2006-01"

	p := journal.Period{}
	s := string(content)

	// Try to cast to a budget date
	date, err := time.Parse(budgetDateFormat, s)
	if err != nil {
		return p, nil
	}

	p.StartDate = date
	p.EndDate = date
	p.Interval = journal.PNone

	return p, nil
}
