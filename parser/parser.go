package parser

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rikchilvers/gledger/journal"
)

// TransactionHandler is the func commands can use to analyse the journal.
// Takes a Transaction and a path to the file where this Transaction was found.
type TransactionHandler = func(t *journal.Transaction, path string) error
type itemParser = func(t itemType, content []rune) error

// Parser is how gledger reads journal files
type Parser struct {
	previousItemType   itemType
	currentPosting     *journal.Posting
	currentTransaction *journal.Transaction
	transactionHandler TransactionHandler
	journalFiles       []string
}

// NewParser creates a parser (including its journal)
func NewParser(t TransactionHandler) Parser {
	return Parser{
		previousItemType:   -1,
		currentPosting:     nil,
		currentTransaction: nil,
		transactionHandler: t,
		journalFiles:       make([]string, 0, 2),
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
		if err := p.endTransaction(); err != nil {
			return err
		}
	case eofItem:
		// Make sure we close the final transaction
		if err := p.endTransaction(); err != nil {
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
		err := p.endTransaction()
		if err != nil {
			return err
		}
		p.currentTransaction = journal.NewTransaction()

		p.currentTransaction.Date, err = parseDate(content)
		if err != nil {
			return fmt.Errorf("error parsing date\n%w", err)
		}
	case stateItem:
		if p.previousItemType != dateItem {
			return fmt.Errorf("expected state but got %s", t)
		}

		switch content[0] {
		case '!':
			p.currentTransaction.State = journal.UnclearedState
		case '*':
			p.currentTransaction.State = journal.ClearedState
		default:
			p.currentTransaction.State = journal.NoState
		}
	case payeeItem:
		if p.previousItemType != dateItem && p.previousItemType != stateItem {
			return fmt.Errorf("expected payee but got %s", t)
		}

		// TODO: try to remove necessity of TrimSpace everywhere
		p.currentTransaction.Payee = strings.TrimSpace(string(content))
	case accountItem:
		if p.previousItemType != payeeItem && p.previousItemType != amountItem && p.previousItemType != accountItem {
			return fmt.Errorf("expected account but got %s", t)
		}

		// Accounts start a posting, so check if we need to start a new one
		// (When a transaction is started, the current posting is set to nil)
		if p.currentPosting != nil {
			p.currentTransaction.AddPosting(p.currentPosting)
		}
		p.currentPosting = journal.NewPosting()

		p.currentPosting.Transaction = p.currentTransaction

		p.currentPosting.AccountPath = string(content)
	case commodityItem:
		if p.previousItemType != accountItem {
			return fmt.Errorf("expected currency but got %s", t)
		}

		if p.currentPosting.Amount == nil {
			p.currentPosting.Amount = journal.NewAmount(string(content), 0)
		} else {
			p.currentPosting.Amount.Commodity = string(content)
		}
	case amountItem:
		if p.previousItemType != commodityItem && p.previousItemType != payeeItem {
			return fmt.Errorf("expected amount but got %s", t)
		}

		if p.currentPosting.Amount == nil {
			p.currentPosting.Amount = journal.NewAmount("", 0)
		}

		if err := p.parseAmount(content); err != nil {
			return fmt.Errorf("error parsing amount: %w", err)
		}
	case periodItem:
		fmt.Println("got a period:", string(content))
	default:
		return fmt.Errorf("unhandled itemType: %s", t)
	}

	p.previousItemType = t
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

func (p *Parser) parseAmount(content []rune) error {
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
			return err
		}
	} else {
		whole, err = strconv.ParseInt(string(content[:decimalPosition]), 10, 64)
		if err != nil {
			return err
		}
		decimal, err = strconv.ParseInt(string(content[decimalPosition+1:]), 10, 64)
		if err != nil {
			return err
		}
	}

	p.currentPosting.Amount.Quantity = multiplier * (whole*100 + decimal)

	return nil
}

func (p *Parser) endTransaction() error {
	if p.currentTransaction == nil {
		return nil
	}

	var err error

	// Make sure we add the last open posting
	if p.currentPosting != nil {
		err = p.currentTransaction.AddPosting(p.currentPosting)
		if err != nil {
			return err
		}
	}

	err = p.currentTransaction.Close()
	if err != nil {
		return err
	}

	if p.transactionHandler != nil {
		if err = p.transactionHandler(p.currentTransaction, p.journalFiles[len(p.journalFiles)-1]); err != nil {
			return err
		}
	}

	p.currentTransaction = nil
	p.currentPosting = nil
	return nil
}
