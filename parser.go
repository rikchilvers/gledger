package main

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

type journalParser interface {
	parseItem(t itemType, content []rune)
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

func (p *parser) parseItem(t itemType, content []rune) {
	switch t {
	case tDate:
		// This will start a transaction so check if we need to close a previous one
		p.endTransaction()
		p.currentTransaction = newTransaction()
		p.currentPosting = nil

		p.currentTransaction.date = parseDate(content)
	case tState:
		if p.previousItemType != tDate {
			log.Fatalln("Unexpected state", p.previousItemType)
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
			log.Fatalln("Unexpected payee", p.previousItemType)
		}

		p.currentTransaction.payee = strings.TrimSpace(string(content))
	case tAccount:
		if p.previousItemType != tPayee && p.previousItemType != tAmount && p.previousItemType != tAccount {
			log.Fatalln("Unexpected account", p.previousItemType)
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
			log.Fatalln("Unexpected currency", p.previousItemType)
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount(0)
		}

		p.currentPosting.amount.commodity = string(content)
	case tAmount:
		if p.previousItemType != tCommodity && p.previousItemType != tPayee {
			log.Fatalln("Unexpected amount", p.previousItemType)
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount(0)
		}

		p.parseAmount(content)
	default:
		fmt.Println("Unhandled itemType", p.previousItemType)
	}

	p.previousItemType = t
}

func parseDate(content []rune) time.Time {
	const DashDateFormat string = "2006-01-02"
	const DotDateFormat string = "2006.01.02"
	const SlashDateFormat string = "2006/01/02"

	var date time.Time
	var err error

	switch content[4] {
	case '-':
		date, err = time.Parse(DashDateFormat, string(content))
	case '.':
		date, err = time.Parse(DotDateFormat, string(content))
	case '/':
		date, err = time.Parse(SlashDateFormat, string(content))
	default:
		log.Fatalln("Malformed date")
	}

	if err != nil {
		log.Fatalln(err)
	}

	return date
}

func (p *parser) parseAmount(content []rune) {
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
	if decimalPosition == -1 {
		// TODO: consider https://stackoverflow.com/a/29255836
		whole, _ = strconv.ParseInt(string(content), 10, 64)
		decimal = 0
	} else {
		whole, _ = strconv.ParseInt(string(content[:decimalPosition]), 10, 64)
		decimal, _ = strconv.ParseInt(string(content[decimalPosition+1:]), 10, 64)
	}

	p.currentPosting.amount.quantity = multiplier * (whole*100 + decimal)
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
