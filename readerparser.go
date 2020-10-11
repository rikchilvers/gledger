package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type ReaderParser struct {
	state              ParserState
	reader             *bufio.Reader
	line               int
	column             int
	currentLine        []rune
	currentTransaction *Transaction
	transactions       []*Transaction
}

func NewReaderParser() *ReaderParser {
	return &ReaderParser{
		state:              AwaitingTransaction,
		reader:             nil,
		line:               1,
		column:             1,
		currentLine:        make([]rune, MaxLineLength),
		currentTransaction: nil,
		transactions:       []*Transaction{},
	}
}

func (p *ReaderParser) NextState() {

}

func (p *ReaderParser) Parse(ledgerFile string) error {
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

		// Skip lines with no content
		if len(p.currentLine) == 0 {
			continue
		}

		// Process the line
		switch p.state {
		case AwaitingTransaction:
			first := p.currentLine[0]
			if isComment(first) {
				// We can skip whole lines of comments
				continue
			} else if isNumeric(first) {
				// We're expecting a transaction header here
				p.parseTransactionHeader()
			} else {
				log.Fatalln("Unexpected character beginning line", p.line)
			}
		default:
			fmt.Println("default state")
		}

		// Handle Reader errors here
	}
	return nil
}

// Reads until the end of a line
// Returns a hint as to whether there are more lines to read
func (p *ReaderParser) ReadLine() bool {
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

func (p *ReaderParser) advanceCaret(n int) {
	p.column += n
	p.currentLine = p.currentLine[n:]
}

// Consume whitespace
// Advances the caret len(whitespace) places
func (p *ReaderParser) consumeWhitespace() {
	fmt.Println(">> consumeWhitespace")
	n := 0
	for {
		r := p.currentLine[n]
		if r == space || r == tab {
			n++
		} else {
			p.advanceCaret(n)
			return
		}
	}
}

// Attemps to parse a comment
// Does not need to advance the caret because comments always end at line end
func (p *ReaderParser) parseComment() (string, error) {
	fmt.Println(">> parseComment")
	start := 0
	if isComment(p.currentLine[0]) {
		start = 1
	}
	comment := p.currentLine[start:]
	return strings.TrimSpace(string(comment)), nil
}

func (p *ReaderParser) parseTransactionHeader() (time.Time, TransactionState, string, string, error) {
	fmt.Println(">> parseTransactionHeader")

	// Parse the date
	date, err := p.parseDate()
	if err != nil {
		return time.Time{}, NoState, "", "", err
	}

	// Handle possible state
	p.consumeWhitespace()
	state := NoState
	if isState(p.currentLine[0]) {
		state = toState(p.currentLine[0])
		p.advanceCaret(1)
	}

	// Handle payee
	p.consumeWhitespace()
	payee, err := p.parsePayee()
	if err != nil {
		return time.Time{}, NoState, "", "", err
	}

	// Handle trailing comment
	comment := ""
	if len(p.currentLine) > 0 {
		p.consumeWhitespace()
		c, err := p.parseComment()
		comment = c
		if err != nil {
			return time.Time{}, NoState, "", "", err
		}
	}

	return date, state, payee, comment, nil
}

// Advances 10 runes to parse the date
func (p *ReaderParser) parseDate() (time.Time, error) {
	fmt.Println(">> parseDate")
	runes := p.currentLine[:10]
	date, err := time.Parse(DateFormat, string(runes))
	if err != nil {
		return time.Time{}, err
	}

	p.advanceCaret(10)
	return date, nil
}

// Advances len(payee)
func (p *ReaderParser) parsePayee() (string, error) {
	fmt.Println(">> parsePayee")

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
