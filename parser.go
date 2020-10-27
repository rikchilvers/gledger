package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type journalParser interface {
	parseItem(t itemType, content []rune) error
}

type parser struct {
	previousItemType   itemType
	currentPosting     *posting
	currentTransaction *transaction
	journal            *journal
}

func newParser() *parser {
	return &parser{
		previousItemType:   -1,
		currentPosting:     nil,
		currentTransaction: nil,
		journal:            newJournal(),
	}
}

func (p *parser) parse(journalPath string) error {
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

func (p *parser) parseItem(t itemType, content []rune) error {
	switch t {
	case tEmptyLine:
		err := p.endTransaction()
		if err != nil {
			return err
		}
	case tEOF:
		// Make sure we close the final transaction
		err := p.endTransaction()
		if err != nil {
			return err
		}
	case tDate:
		// This will start a transaction so check if we need to close a previous one
		err := p.endTransaction()
		if err != nil {
			return err
		}
		p.currentTransaction = newTransaction()

		p.currentTransaction.date, err = parseDate(content)
		if err != nil {
			return fmt.Errorf("Error parsing date\n%w", err)
		}
	case tState:
		if p.previousItemType != tDate {
			return fmt.Errorf("Expected state but got %s", t)
		}

		switch content[0] {
		case '!':
			p.currentTransaction.state = tUncleared
		case '*':
			p.currentTransaction.state = tCleared
		default:
			p.currentTransaction.state = tNoState
		}
	case tPayee:
		if p.previousItemType != tDate && p.previousItemType != tState {
			return fmt.Errorf("Expected payee but got %s", t)
		}

		p.currentTransaction.payee = strings.TrimSpace(string(content))
	case tAccount:
		if p.previousItemType != tPayee && p.previousItemType != tAmount && p.previousItemType != tAccount {
			return fmt.Errorf("Expected account but got %s", t)
		}

		// Accounts start a posting, so check if we need to start a new one
		// (When a transaction is started, the current posting is set to nil)
		if p.currentPosting != nil {
			p.currentTransaction.addPosting(p.currentPosting)
		}
		p.currentPosting = newPosting()

		p.currentPosting.transaction = p.currentTransaction

		// TODO: try to remove necessity of TrimSpace everywhere
		// a := strings.TrimSpace(string(content))
		// p.currentPosting.account = newAccountWithChildren(strings.Split(a, ":"), nil)
		p.currentPosting.accountPath = strings.Split(strings.TrimSpace(string(content)), ":")
	case tCommodity:
		if p.previousItemType != tAccount {
			return fmt.Errorf("Expected currency but got %s", t)
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount(0)
		}

		p.currentPosting.amount.commodity = string(content)
	case tAmount:
		if p.previousItemType != tCommodity && p.previousItemType != tPayee {
			return fmt.Errorf("Expected amount but got %s", t)
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount(0)
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
	const DotDateFormat string = "2006.01.02"
	const SlashDateFormat string = "2006/01/02"

	s := string(content)
	var date time.Time
	var err error

	switch content[4] {
	case '-':
		date, err = time.Parse(DashDateFormat, s)
	case '.':
		date, err = time.Parse(DotDateFormat, s)
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

func (p *parser) parseAmount(content []rune) error {
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

	p.currentPosting.amount.quantity = multiplier * (whole*100 + decimal)

	return nil
}

func (p *parser) endTransaction() error {
	if p.currentTransaction == nil {
		return nil
	}

	var err error

	// Make sure we add the last open posting
	if p.currentPosting != nil {
		err = p.currentTransaction.addPosting(p.currentPosting)
		if err != nil {
			return err
		}
	}

	err = p.currentTransaction.close()
	if err != nil {
		return err
	}
	p.journal.addTransaction(p.currentTransaction)

	p.currentTransaction = nil
	p.currentPosting = nil
	return nil
}
