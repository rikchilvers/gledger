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
				fmt.Println("comment is:", string(p.currentLine))
			} else if isNumeric(first) {
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

func (p *ReaderParser) parseTransactionHeader() (time.Time, TransactionState, string, error) {
	// Parse the date
	date, err := p.parseDate()
	if err != nil {
		return time.Time{}, NoState, "", err
	}

	return date, NoState, "", nil
}

func (p *ReaderParser) advanceCaret(n int) {
	p.column += n
	p.currentLine = p.currentLine[n:]
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

// Consumes all whitespace until a newline or non-whitespace
// Consumes newlines before returning
// Returns a hint as to whether the nonwhitespace character is a newline or not
// func (p *ReaderParser) consumeWhitespace() bool {
// 	fmt.Println(">> consumeWhitespace")
// 	next := p.advance()
// 	for {
// 		if next == space || next == tab {
// 			next = p.advance()
// 		} else if next == newline || next == carriageReturn {
// 			return true
// 		} else {
// 			p.reader.UnreadRune()
// 			return false
// 		}
// 	}
// }

// Attemps to parse a comment
// func (p *ReaderParser) parseComment() {
// 	fmt.Println(">> parseComment")
// 	// Collect the comment here
// 	runes := []rune{}

// 	next := p.advance()
// 	if !isComment(next) {
// 		p.reader.UnreadRune()
// 		return
// 	}

// 	// Read until a new line
// 	for {
// 		if isNewline(next) {
// 			break
// 		}
// 		// runes = append(runes, r)
// 	}

// 	comment := string(runes)
// 	fmt.Println("The comment is:", comment)
// }
