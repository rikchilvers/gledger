package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"
)

// The date format journal files use
const DateFormat string = "2006-01-02"

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
		// This will start a transaction
		// Check if we need to close a previous one
		p.endTransaction()
		p.currentTransaction = newTransaction()
		p.currentPosting = nil

		date, err := time.Parse(DateFormat, string(content))
		if err != nil {
			log.Fatalln(err)
		}
		p.currentTransaction.date = date
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

		p.currentPosting.account = newAccount(strings.TrimSpace(string(content)))
	case tCommodity:
		if p.previousItemType != tAccount {
			log.Fatalln("Unexpected currency", p.previousItemType)
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount()
		}

		p.currentPosting.amount.commodity = strings.TrimSpace(string(content))
	case tAmount:
		if p.previousItemType != tCommodity && p.previousItemType != tPayee {
			log.Fatalln("Unexpected amount", p.previousItemType)
		}

		if p.currentPosting.amount == nil {
			p.currentPosting.amount = newAmount()
		}

		p.currentPosting.amount.quantity = int64(1)
	default:
		fmt.Println("Unhandled itemType", p.previousItemType)
	}

	p.previousItemType = t
}

func (p *parser) endTransaction() {
	if p.currentTransaction == nil {
		return
	}

	// Make sure we add the last open posting
	// TODO: do this at the end of the file too
	if p.currentPosting != nil {
		p.currentTransaction.addPosting(p.currentPosting)
	}
	p.currentTransaction.close()
	p.journal.transactions = append(p.journal.transactions, p.currentTransaction)
}
