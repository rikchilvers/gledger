package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type Parser struct {
	state              ParserState
	reader             *bufio.Reader
	line               int
	column             int
	currentLine        []rune
	currentTransaction *Transaction
	transactions       []*Transaction
}

func NewParser() *Parser {
	return &Parser{
		state:              AwaitingTransaction,
		reader:             nil,
		line:               1,
		column:             1,
		currentLine:        make([]rune, MaxLineLength),
		currentTransaction: nil,
		transactions:       []*Transaction{},
	}
}

func (p *Parser) NextState() {
	switch p.state {
	case AwaitingTransaction:
		p.state = InTransaction
	case InTransaction:
		p.state = AwaitingTransaction
		// TODO: close transaction
		// TODO: add transaction to ledger
	default:
		return
	}
}

func (p *Parser) Parse(ledgerFile string) error {
	fmt.Println(">> parse", ledgerFile)
	file, err := os.Open(ledgerFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	p.reader = bufio.NewReader(file)

	for {
		// Read a line - break if we can't read anymore
		if !p.ReadLine() {
			fmt.Printf("there are %d lines in the file\n", p.line)
			break
		}

		// Process the line
		switch p.state {
		case AwaitingTransaction:
			// We can skip commented lines and lines with no content
			if len(p.currentLine) == 0 || isComment(p.currentLine[0]) {
				continue
			} else if isNumeric(p.currentLine[0]) {
				// We're expecting a transaction header here
				p.parseTransactionHeader()
				fmt.Printf("%+v\n", p.currentTransaction)
			} else {
				log.Fatalln("Unexpected character beginning line", p.line)
			}

			p.NextState()
		case InTransaction:
			// Lines with no content are interpreted as closing the transaction
			if len(p.currentLine) == 0 {
				p.NextState()
				continue
			}

			// Make sure the line starts with enough whitespace
			if p.consumeWhitespace() < 2 {
				log.Fatalln("not enough spaces before posting account/comment")
			}

			// At this point, we're expecting an account line or a comment
			if isComment(p.currentLine[0]) {
				p.parseComment()
			} else {
				p.parsePosting()
			}
			// TODO add posting to transaction
		case Stop:
			return nil
		default:
			fmt.Println("default state")
		}

		// Handle Reader errors here
	}
	return nil
}

// Reads until the end of a line
//
// Returns a hint as to whether there are more lines to read
func (p *Parser) ReadLine() bool {
	// fmt.Println(">> ReadLine")
	// Reset the line
	p.currentLine = nil
	for {
		r, _, err := p.reader.ReadRune()
		if err != nil {
			fmt.Println("!! Error reading file:", err)
			return false
		}

		if r == newline || r == carriageReturn {
			p.line++
			p.column = 1
			return true
		} else {
			p.currentLine = append(p.currentLine, r)
		}
	}
}

func (p *Parser) advanceCaret(n int) {
	p.column += n
	p.currentLine = p.currentLine[n:]
}

// Consume whitespace
//
// Advances the caret len(whitespace) places and returns this count
func (p *Parser) consumeWhitespace() int {
	fmt.Println(">> consumeWhitespace on line:", p.line)
	n := 0
	spaces := 0
	for {
		r := p.currentLine[n]
		if r == space {
			n++
			spaces += 1
		} else if r == tab {
			n++
			spaces += SpacesPerTab
		} else {
			p.advanceCaret(n)
			return spaces
		}
	}
}

// Attemps to parse a comment
// Does not need to advance the caret because comments always end at line end
func (p *Parser) parseComment() (string, error) {
	fmt.Println(">> parseComment on line:", p.line)
	start := 0
	if isComment(p.currentLine[0]) {
		start = 1
	}
	comment := p.currentLine[start:]
	return strings.TrimSpace(string(comment)), nil
}

func (p *Parser) parseTransactionHeader() error {
	fmt.Println(">> parseTransactionHeader on line:", p.line)
	p.currentTransaction = &Transaction{}

	// Parse the date
	date, err := p.parseDate()
	if err != nil {
		return err
	}
	p.currentTransaction.date = date

	// Handle possible state
	p.consumeWhitespace()
	state := NoState
	if isState(p.currentLine[0]) {
		state = toState(p.currentLine[0])
		p.advanceCaret(1)
	}
	p.currentTransaction.state = state

	// Handle payee
	p.consumeWhitespace()
	payee, err := p.parsePayee()
	if err != nil {
		return err
	}
	p.currentTransaction.payee = payee

	// Handle trailing comment
	// comment := ""
	// if len(p.currentLine) > 0 {
	// 	p.consumeWhitespace()
	// 	c, err := p.parseComment()
	// 	comment = c
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

func (p *Parser) parsePosting() error {
	fmt.Println(">> parsePosting on line:", p.line)
	account, err := p.parseAccount()
	if err != nil {
		return err
	}

	if len(p.currentLine) == 0 {
		fmt.Println("\telided amount")
		return nil
	}

	if p.consumeWhitespace() < 2 {
		log.Fatalln("not enough spaces before posting amount on line:", p.line)
	}

	currency, err := p.parseCurrency()
	amount, err := p.parseAmount()

	// Construct the posting
	posting := newPosting(account, currency, amount)
	p.currentTransaction.addPosting(posting)

	// There could be a comment at the end of the line
	if len(p.currentLine) > 0 {
		p.consumeWhitespace()
		p.parseComment()
	}

	return nil
}

// Advances 10 runes to parse the date
func (p *Parser) parseDate() (time.Time, error) {
	fmt.Println(">> parseDate on line:", p.line)
	runes := p.currentLine[:10]
	date, err := time.Parse(DateFormat, string(runes))
	if err != nil {
		return time.Time{}, err
	}

	p.advanceCaret(10)
	return date, nil
}

// Advances len(payee)
func (p *Parser) parsePayee() (string, error) {
	fmt.Println(">> parsePayee on line:", p.line)

	n := 0
	// We need to go up to the end of the line or the start of a comment
	for _, r := range p.currentLine {
		if isComment(r) {
			break
		} else {
			n++
		}
	}

	payee := p.currentLine[:n]
	p.advanceCaret(n)

	return strings.TrimSpace(string(payee)), nil
}

// Parses an account
func (p *Parser) parseAccount() (string, error) {
	fmt.Println(">> parseAccount on line:", p.line)

	// Search for index of >= 2 spaces
	spaces := 0
	firstSpaceIndex := -1
	for i, r := range p.currentLine {
		switch r {
		case space:
			if spaces == 0 {
				firstSpaceIndex = i
			}
			spaces += 1
		case tab:
			if spaces == 0 {
				firstSpaceIndex = i
			}
			spaces += SpacesPerTab
		default:
			spaces = 0
			firstSpaceIndex = -1
		}

		if spaces > 1 {
			break
		}
	}

	// Take up to the end of the account
	account := ""
	if firstSpaceIndex != -1 {
		account = string(p.currentLine[:firstSpaceIndex])
	} else {
		account = string(p.currentLine)
	}

	p.advanceCaret(firstSpaceIndex + 1)

	return account, nil
}

func (p *Parser) parseCurrency() (interface{}, error) {
	fmt.Println(">> parseCurrency on line:", p.line)

	fmt.Println("\t", string(p.currentLine))

	if isNumeric(p.currentLine[0]) {
		return nil, nil
	}

	n := 0
	for i, r := range p.currentLine {
		if isNumeric(r) {
			n = i
			break
		}
	}

	currency := p.currentLine[:n]
	fmt.Println("\tcurrency is", string(currency))
	p.advanceCaret(n)

	return currency, nil
}

func (p *Parser) parseAmount() (interface{}, error) {
	fmt.Println(">> parseAmount on line:", p.line)

	fmt.Println("\t", string(p.currentLine))

	// TODO: actually read the amount

	return nil, nil
}
