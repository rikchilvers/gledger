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
	zero           = rune('0')
	nine           = rune('9')
	tab            = rune('\t')
	newline        = rune('\n')
	carriageReturn = rune('\r')
	space          = rune(' ')
	semicolon      = rune(';')
	forwardSlash   = rune('/')
	hash           = rune('#')
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

func (p *ReaderParser) Parse(ledgerFile string) {
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
				fmt.Println("comment is:", string(p.currentLine))
			} else if isNumeric(first) {
				p.parseDate()
			} else {
				log.Fatalln("Unexpected character beginning line", p.currentLine)
			}
		default:
			fmt.Println("default state")
		}

		// Handle Reader errors here
	}

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
			return true
		} else {
			p.currentLine = append(p.currentLine, r)
		}
	}
}

func (p *ReaderParser) peek(n int) rune {
	fmt.Println(">> peek", n)
	b, err := p.reader.Peek(n)
	if err != nil {
		fmt.Println("\nError reading file:", err)
		p.state = Stop
	}
	return rune(b[0])
}

func (p *ReaderParser) advance() rune {
	r, _, err := p.reader.ReadRune()
	if err != nil {
		fmt.Println("\nError reading file:", err)
		p.state = Stop
	}

	if r == newline || r == carriageReturn {
		p.line++
	}

	fmt.Printf(">> advanced to ln %d, col %d\n", p.line, p.column)

	return r
}

// Advances 10 runes to parse the date
func (p *ReaderParser) parseDate() {
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

// Consumes all whitespace until a newline or non-whitespace
// Consumes newlines before returning
// Returns a hint as to whether the nonwhitespace character is a newline or not
func (p *ReaderParser) consumeWhitespace() bool {
	fmt.Println(">> consumeWhitespace")
	next := p.advance()
	for {
		if next == space || next == tab {
			next = p.advance()
		} else if next == newline || next == carriageReturn {
			return true
		} else {
			p.reader.UnreadRune()
			return false
		}
	}
}

// Attemps to parse a comment
func (p *ReaderParser) parseComment() {
	fmt.Println(">> parseComment")
	// Collect the comment here
	runes := []rune{}

	next := p.advance()
	if !isComment(next) {
		p.reader.UnreadRune()
		return
	}

	// Read until a new line
	for {
		if isNewline(next) {
			break
		}
		// runes = append(runes, r)
	}

	comment := string(runes)
	fmt.Println("The comment is:", comment)
}
