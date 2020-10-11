package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"
)

// Runes
const (
	zero         = rune('0')
	nine         = rune('9')
	tab          = rune('\t')
	newline      = rune('\n')
	space        = rune(' ')
	semicolon    = rune(';')
	forwardSlash = rune('/')
	hash         = rune('#')
)

const DateFormat string = "2006-01-02"

type ParserState int

const (
	AwaitingTransaction ParserState = iota
	ExpectingDate
	Stop
)

type Parser struct {
	state              ParserState
	reader             *bufio.Reader
	line               int
	column             int
	currentTransaction *Transaction
	transactions       []*Transaction
}

func NewParser() *Parser {
	return &Parser{
		state:              AwaitingTransaction,
		reader:             nil,
		line:               0,
		column:             0,
		currentTransaction: nil,
		transactions:       []*Transaction{},
	}
}

func (p *Parser) Parse(ledgerFile string) {
	fmt.Println(">> parse", ledgerFile)
	file, err := os.Open(ledgerFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	p.reader = bufio.NewReader(file)

	// Consume whitespace
	p.consumeWhitespace()

	// Check for comments at the start of a line
	r := p.peek(1)
	fmt.Println(rune('\n'))
	if beginComment(r) {
		p.parseComment()
	}

	for {
		switch p.state {
		case AwaitingTransaction:
			p.consumeWhitespace()
			p.parseDate()
		case Stop:
			break
		default:
			fmt.Println("default state")
		}
	}
}

func (p *Parser) peek(n int) rune {
	fmt.Println(">> peek", n)
	b, err := p.reader.Peek(n)
	if err != nil {
		fmt.Println("\nError reading file:", err)
		p.state = Stop
	}
	return rune(b[0])
}

func (p *Parser) advance() rune {
	r, _, err := p.reader.ReadRune()
	if err != nil {
		fmt.Println("\nError reading file:", err)
		p.state = Stop
	}

	if r == newline {
		p.line++
		p.column = 0
	} else {
		p.column++
	}

	fmt.Printf(">> advanced to ln %d, col %d\n", p.line, p.column)

	return r
}

// Advances 10 runes to parse the date
func (p *Parser) parseDate() {
	fmt.Println(">> parseDate")
	runes := []rune{}
	for i := 0; i < 10; i++ {
		runes = append(runes, p.advance())
	}
	date, err := time.Parse(DateFormat, string(runes))
	if err != nil {
		log.Fatal("Failed to parse date")
	}

	fmt.Println("The date is:", date)
}

func beginComment(r rune) bool {
	return r == semicolon || r == hash
}

func (p *Parser) consumeWhitespace() {
	fmt.Println(">> consumeWhitespace")
	next := p.advance()
	for {
		if next == space || next == tab || next == newline {
			next = p.advance()
		} else {
			p.reader.UnreadRune()
			break
		}
	}
}

func (p *Parser) parseComment() {
	fmt.Println(">> parseComment")
	runes := []rune{}
	for {
		r := p.advance()
		if r == newline {
			break
		}
		runes = append(runes, r)
	}

	comment := string(runes)
	fmt.Println("The comment is:", comment)
}
