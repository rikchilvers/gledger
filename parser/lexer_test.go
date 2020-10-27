package parser

import (
	"strings"
	"testing"
)

const transactionHeaderWithComment string = "2020-10-11 A shop"

var transactionHeader []byte = []byte("2020-10-11 A shop")
var transactionDate string = "2020-10-11"
var transactionPayee string = "A shop"

type mockParser struct {
	lexedDate  bool
	lexedState bool
	lexedPayee bool
}

func (p *mockParser) parseItem(t itemType, content []rune) error {
	switch t {
	case dateItem:
		if string(content) == transactionDate {
			p.lexedDate = true
		}
	case stateItem:
		p.lexedState = true
	case payeeItem:
		if string(content) == transactionPayee {
			p.lexedPayee = true
		}
	default:
		return nil
	}

	return nil
}

func TestLex(t *testing.T) {
	parser := &mockParser{}
	reader := strings.NewReader(string(transactionHeader))
	lexer := lexer{}
	lexer.currentLine = 1
	lexer.parser = parser

	lexer.lex(reader)
}

func TestLexTransactionHeader(t *testing.T) {
	parser := &mockParser{}
	lexer := lexer{}
	lexer.currentLine = 1
	lexer.input = transactionHeader
	lexer.parser = parser

	lexer.lexTransactionHeader()

	if !parser.lexedDate {
		t.Fatalf("lexer did not lex date")
	}

	if !parser.lexedPayee {
		t.Fatalf("lexer did not lex payee")
	}

	if parser.lexedState {
		t.Fatalf("lexer erroneously lexed state")
	}
}

func TestLexTransactionHeaderWithComment(t *testing.T) {

}
