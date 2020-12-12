package parser

import (
	"strings"
	"testing"
)

const transactionHeaderWithComment string = "2020-10-11 A shop"

var (
	transactionHeader []byte = []byte("2020-10-11 A shop")
	transactionDate   string = "2020-10-11"
	transactionPayee  string = "A shop"
)

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

func TestTakeToNextLineOrComment(t *testing.T) {
	expected := "input test"
	input := strings.NewReader(string("input test\n"))
	lexer := newLexer(input, "test", nil)
	if err := lexer.ingest(); err != nil {
		t.Fatalf("failed to ingest")
	}
	output := string(lexer.takeToTabOrNextLineOrComment())
	if output != expected {
		t.Fatalf("newline fail expected '%s' got '%s'", expected, output)
	}

	input = strings.NewReader(string("input test;comment"))
	lexer = newLexer(input, "test", nil)
	if err := lexer.ingest(); err != nil {
		t.Fatalf("failed to ingest")
	}
	output = string(lexer.takeToTabOrNextLineOrComment())
	if output != expected {
		t.Fatalf("comment fail - expected '%s' got '%s'", expected, output)
	}

	input = strings.NewReader(string("input test  "))
	lexer = newLexer(input, "test", nil)
	if err := lexer.ingest(); err != nil {
		t.Fatalf("failed to ingest")
	}
	output = string(lexer.takeToTabOrNextLineOrComment())
	if output != expected {
		t.Fatalf("spaces fail - expected '%s' got '%s'", expected, output)
	}

	input = strings.NewReader(string("input test "))
	lexer = newLexer(input, "test", nil)
	if err := lexer.ingest(); err != nil {
		t.Fatalf("failed to ingest")
	}
	output = string(lexer.takeToTabOrNextLineOrComment())
	if output != expected {
		t.Fatalf("single space fail - expected '%s' got '%s'", expected, output)
	}

	input = strings.NewReader(string("input test\thello"))
	lexer = newLexer(input, "test", nil)
	if err := lexer.ingest(); err != nil {
		t.Fatalf("failed to ingest")
	}
	output = string(lexer.takeToTabOrNextLineOrComment())
	if output != expected {
		t.Fatalf("tab fail - expected '%s' got '%s'", expected, output)
	}
}

// TestLex checks lexing a reader works
// (this is how the lexer will be used in production)
func TestLex(t *testing.T) {
	parser := mockParser{}
	lexer := newLexer(strings.NewReader(string(transactionHeader)), "th", parser.parseItem)
	err := lexer.lex()
	if err != nil {
		t.Fatalf("lex returned unexpected err")
	}
}

func TestLexTransactionHeader(t *testing.T) {
	parser := mockParser{}
	lexer := lexer{}
	lexer.parser = parser.parseItem
	lexer.currentLine = 1
	lexer.input = transactionHeader

	err := lexer.lexTransactionHeader()
	if err != nil {
		t.Fatalf("lex returned unexpected err")
	}

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
