package parser

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	. "github.com/rikchilvers/gledger/journal"
)

type journalParser interface {
	parseItem(t itemType, content []rune) error
}

// Parser is how gledger reads journal files
type Parser struct {
	previousItemType   itemType
	currentPosting     *Posting
	currentTransaction *Transaction
	journal            *Journal
}

// NewParser creates a parser (including its journal)
func NewParser() *Parser {
	return &Parser{
		previousItemType:   -1,
		currentPosting:     nil,
		currentTransaction: nil,
		journal:            NewJournal(),
	}
}

// Parse lexes and parses the provided file line by line
func (p *Parser) Parse(journalPath string) error {
	file, err := os.Open(journalPath)
	if err != nil {
		return err
	}
	defer file.Close()

	lexer := lexer{}
	lexer.parser = p

	err = lexer.lex(file)
	if err != nil {
		// This is the exit point for the lexer's errors
		return fmt.Errorf("Error at %s%w", journalPath, err)
	}

	return nil
}

func (p *Parser) parseItem(t itemType, content []rune) error {
	switch t {
	case emptyLineItem:
		err := p.endTransaction()
		if err != nil {
			return err
		}
	case eofItem:
		// Make sure we close the final transaction
		err := p.endTransaction()
		if err != nil {
			return err
		}
	case dateItem:
		// This will start a transaction so check if we need to close a previous one
		err := p.endTransaction()
		if err != nil {
			return err
		}
		p.currentTransaction = NewTransaction()

		p.currentTransaction.Date, err = parseDate(content)
		if err != nil {
			return fmt.Errorf("Error parsing date\n%w", err)
		}
	case stateItem:
		if p.previousItemType != dateItem {
			return fmt.Errorf("Expected state but got %s", t)
		}

		switch content[0] {
		case '!':
			p.currentTransaction.State = UnclearedState
		case '*':
			p.currentTransaction.State = ClearedState
		default:
			p.currentTransaction.State = NoState
		}
	case payeeItem:
		if p.previousItemType != dateItem && p.previousItemType != stateItem {
			return fmt.Errorf("Expected payee but got %s", t)
		}

		p.currentTransaction.Payee = strings.TrimSpace(string(content))
	case accountItem:
		if p.previousItemType != payeeItem && p.previousItemType != amountItem && p.previousItemType != accountItem {
			return fmt.Errorf("Expected account but got %s", t)
		}

		// Accounts start a posting, so check if we need to start a new one
		// (When a transaction is started, the current posting is set to nil)
		if p.currentPosting != nil {
			p.currentTransaction.AddPosting(p.currentPosting)
		}
		p.currentPosting = NewPosting()

		p.currentPosting.Transaction = p.currentTransaction

		// TODO: try to remove necessity of TrimSpace everywhere
		// a := strings.TrimSpace(string(content))
		// p.currentPosting.account = newAccountWithChildren(strings.Split(a, ":"), nil)
		p.currentPosting.AccountPath = strings.Split(strings.TrimSpace(string(content)), ":")
	case commodityItem:
		if p.previousItemType != accountItem {
			return fmt.Errorf("Expected currency but got %s", t)
		}

		if p.currentPosting.Amount == nil {
			p.currentPosting.Amount = NewAmount(0)
		}

		p.currentPosting.Amount.Commodity = string(content)
	case amountItem:
		if p.previousItemType != commodityItem && p.previousItemType != payeeItem {
			return fmt.Errorf("Expected amount but got %s", t)
		}

		if p.currentPosting.Amount == nil {
			p.currentPosting.Amount = NewAmount(0)
		}

		err := p.parseAmount(content)
		if err != nil {
			return fmt.Errorf("error parsing amount: %w", err)
		}
	default:
		return fmt.Errorf("Unhandled itemType: %s", t)
	}

	p.previousItemType = t
	return nil
}

func parseDate(content []rune) (time.Time, error) {
	const DashDateFormat string = "2006-01-02"
	const DodateItemFormat string = "2006.01.02"
	const SlashDateFormat string = "2006/01/02"

	s := string(content)
	var date time.Time
	var err error

	switch content[4] {
	case '-':
		date, err = time.Parse(DashDateFormat, s)
	case '.':
		date, err = time.Parse(DodateItemFormat, s)
	case '/':
		date, err = time.Parse(SlashDateFormat, s)
	default:
		return time.Time{}, fmt.Errorf("Date is malformed: %s", s)
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
	} else {
		whole, err = strconv.ParseInt(string(content[:decimalPosition]), 10, 64)
		decimal, err = strconv.ParseInt(string(content[decimalPosition+1:]), 10, 64)
	}

	if err != nil {
		return err
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
	p.journal.AddTransaction(p.currentTransaction)

	p.currentTransaction = nil
	p.currentPosting = nil
	return nil
}