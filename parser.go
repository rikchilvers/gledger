package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

type journalParser interface {
	parseItem(t itemType, content []rune, line int) error
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

func (p *parser) parse(r io.Reader) {
	lexer := lexer{}
	lexer.parser = p

	lexer.lex(r)

	// Make sure we close the final transaction
	p.endTransaction()

	for _, t := range p.journal.transactions {
		fmt.Println()
		fmt.Println(t)
	}
}

func (p *parser) parseItem(t itemType, content []rune, line int) error {
	var err error

	switch t {
	case tDate:
		// This will start a transaction so check if we need to close a previous one
		p.endTransaction()
		p.currentTransaction = newTransaction()
		p.currentPosting = nil

		p.currentTransaction.date, err = parseDate(content)
		if err != nil {
			return fmt.Errorf("error parsing date: %w", err)
		}
	case tState:
		if p.previousItemType != tDate {
			return errors.New(fmt.Sprintf("expected state but got", p.previousItemType))
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
			return errors.New(fmt.Sprintf("expected payee but got", p.previousItemType))
		}

		p.currentTransaction.payee = strings.TrimSpace(string(content))
	case tAccount:
		if p.previousItemType != tPayee && p.previousItemType != tAmount && p.previousItemType != tAccount {
			return errors.New(fmt.Sprintf("expected account but go", p.previousItemType))
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
			return errors.New(fmt.Sprintf("expected currency but got", p.previousItemType))
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount(0)
		}

		p.currentPosting.amount.commodity = string(content)
	case tAmount:
		if p.previousItemType != tCommodity && p.previousItemType != tPayee {
			return errors.New(fmt.Sprintf("expected amount but got", p.previousItemType))
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount(0)
		}

		err = p.parseAmount(content)
		if err != nil {
			return fmt.Errorf("error parsing amount: %w", err)
		}
	default:
		return errors.New(fmt.Sprintf("Unhandled itemType", p.previousItemType))
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
		return time.Time{}, errors.New(fmt.Sprintf("could not parse malformed date: %s", s))
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

func (p *parser) endTransaction() {
	if p.currentTransaction == nil {
		return
	}

	// Make sure we add the last open posting
	if p.currentPosting != nil {
		p.currentTransaction.addPosting(p.currentPosting)
	}
	p.currentTransaction.close()

	p.journal.addTransaction(p.currentTransaction)
}
