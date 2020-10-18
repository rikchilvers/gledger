package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode"
)

type OldParser struct {
	state       ParserState
	reader      *bufio.Reader
	line        int
	column      int
	currentLine []rune
	// We might need to loop multiple lines for a posting if it has comments, so keep track of it here
	currentPosting     *posting
	currentTransaction *transaction
	transactions       []*transaction
}

func NewParser() *OldParser {
	return &OldParser{
		state:              AwaitingTransaction,
		reader:             nil,
		line:               0, // has to be zero because of how ReadLine works
		column:             1,
		currentLine:        make([]rune, MaxLineLength),
		currentTransaction: nil,
		currentPosting:     nil,
		transactions:       []*transaction{},
	}
}

func (p *OldParser) NextState() {
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

func (p *OldParser) Parse(ledgerFile string) error {
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

		// Determine the type of line

		if len(p.currentLine) == 0 {
			fmt.Println("skip")
			continue
		}

		firstRune := p.currentLine[0]

		if isComment(firstRune) {
			fmt.Println("comment")
		}

		if unicode.IsNumber(firstRune) {
			fmt.Println("transaction header")
		}

		// Other types of first char check go here

		if p.consumeWhitespace() > 2 {
			fmt.Println("posting")
		}

		continue

		// Process the line
		switch p.state {
		case AwaitingTransaction:
			// We can skip commented lines and lines with no content
			if len(p.currentLine) == 0 {
				fmt.Println(".. Skipping empty line", p.line)
				continue
			} else if isComment(p.currentLine[0]) {
				fmt.Println(".. Skipping comment line", p.line)
				continue
			} else if isNumeric(p.currentLine[0]) {
				// We're expecting a transaction header here
				err := p.parseTransactionHeader()
				if err != nil {
					return err
				}
				fmt.Printf("%+v\n", p.currentTransaction)
			} else {
				log.Fatalln("Unexpected character beginning line", p.line)
			}

			p.NextState()
		case InTransaction:
			// Lines with no content close the transaction
			if len(p.currentLine) == 0 {
				p.NextState()
				continue
			}

			// Lines without enough whitespace close the transaction
			if p.consumeWhitespace() < 2 {
				p.NextState()
				continue
			}

			if p.currentPosting == nil {
				p.currentPosting = &posting{}
			}

			// At this point, we're expecting an account line or a comment
			if isComment(p.currentLine[0]) {
				comment, err := p.parseComment()
				if err != nil {
					return err
				}
				if len(comment) > 0 {
					p.currentPosting.comments = append(p.currentPosting.comments, comment)
				}
			} else {
				p.parsePosting()
			}
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
func (p *OldParser) ReadLine() bool {
	fmt.Println(">> ReadLine", p.line+1)
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

func (p *OldParser) advanceCaret(n int) {
	p.column += n
	p.currentLine = p.currentLine[n:]
}

// Consume whitespace
//
// Advances the caret len(whitespace) places and returns this count
func (p *OldParser) consumeWhitespace() int {
	// fmt.Println(">> consumeWhitespace on line:", p.line)
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
func (p *OldParser) parseComment() (string, error) {
	fmt.Println(">> parseComment on line:", p.line)
	start := 0
	if isComment(p.currentLine[0]) {
		start = 1
	}
	comment := p.currentLine[start:]
	return strings.TrimSpace(string(comment)), nil
}

func (p *OldParser) parseTransactionHeader() error {
	fmt.Println(">> parseTransactionHeader on line:", p.line)
	p.currentTransaction = &transaction{}

	// Parse the date
	date, err := p.parseDate()
	if err != nil {
		return err
	}
	p.currentTransaction.date = date

	// Handle possible state
	p.consumeWhitespace()
	state := tNoState
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

func (p *OldParser) parsePosting() error {
	fmt.Println(">> parsePosting on line:", p.line)
	account, err := p.parseAccount()
	if err != nil {
		return err
	}

	// Check if more than 2 runes remain on the line
	fmt.Println("line has chars remaining:", len(p.currentLine))
	if len(p.currentLine) < 2 {
		return nil
	}

	if p.consumeWhitespace() < 2 {
		log.Fatalln("not enough spaces before posting amount on line:", p.line)
	}

	currency, err := p.parseCurrency()
	amount, err := p.parseAmount()

	// Construct the posting
	p.currentPosting.account = account
	p.currentPosting.currency = currency
	p.currentPosting.amount = amount
	p.currentTransaction.addPosting(*p.currentPosting)

	return nil
}

// Advances 10 runes to parse the date
func (p *OldParser) parseDate() (time.Time, error) {
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
func (p *OldParser) parsePayee() (string, error) {
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
func (p *OldParser) parseAccount() (string, error) {
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

	// Take up to the end of the line or the account (if it is followed by > 2 spaces)
	account := ""
	if spaces >= 2 {
		// if we have >=2 spaces, we can take up to the first space index
		account = string(p.currentLine[:firstSpaceIndex])
		// advance to the first space
		p.advanceCaret(firstSpaceIndex + 1)
	} else {
		// otherwise, we just take the whole line
		account = strings.TrimSpace(string(p.currentLine))
		// advance to the end of the line
		p.currentLine = p.currentLine[:0]
	}

	return account, nil
}

func (p *OldParser) parseCurrency() (interface{}, error) {
	fmt.Println(">> parseCurrency on line:", p.line)

	// The currency might be elided
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
	p.advanceCaret(n)

	return currency, nil
}

func (p *OldParser) parseAmount() (interface{}, error) {
	fmt.Println(">> parseAmount on line:", p.line)

	n := -1
	for i, r := range p.currentLine {
		if !(isNumeric(r) || r == period) {
			n = i
			break
		}
	}
	if n == -1 {
		n = len(p.currentLine)
	}

	amount := p.currentLine[:n]
	p.advanceCaret(n)

	return amount, nil
}
