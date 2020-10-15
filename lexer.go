// The lexer reads a file of bytes and identifies components to be parsed

package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"unicode"
	"unicode/utf8"
)

type ItemType int

const (
	tDate ItemType = iota
	tState
	tPayee
	tAccount
	tCurrency
	tAmount
	tComment
)

const EOF = -1
const TabWidth = 2 // size of a tab in spaces

type Lexer struct {
	reader *bufio.Reader
	input  []byte
	pos    int // input position
	start  int // item start position
	width  int // width of last element
}

// Lexes the file line by line
func (l *Lexer) IngestLine(r io.Reader) {
	l.reader = bufio.NewReader(r)

	count := 1
	for {
		line, isPrefix, err := l.reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Fatalln("Error!", err)
		}
		if isPrefix {
			log.Fatalln("Unhandled split line")
		}

		fmt.Printf("Line %2d\n", count)
		// Reset the positions
		l.pos = 0
		l.start = 0
		l.width = 0

		l.input = line
		l.lexLine()
		count++
	}
}

// Lex the line
func (l *Lexer) lexLine() {
	// Bail early if the line is empty
	if len(l.input) == 0 {
		return
	}

	firstRune := l.next()
	if firstRune == EOF {
		return
	}
	secondRune := l.peek()

	// Detect transaction headers
	// TODO: handle passing different transaction header indicators
	if unicode.IsNumber(firstRune) {
		fmt.Println(("\ttransaction header"))
		l.lexTransactionHeader()
		return
	}

	// Detect posting lines
	if firstRune == '\t' || firstRune == ' ' && secondRune == ' ' {
		fmt.Println("\tposting line")
		l.lexPostingLine()
		return
	}

	return
}

func (l *Lexer) lexTransactionHeader() {
	// Need to backup to include the first rune
	l.backup()

	date := l.takeUntilSpace()
	fmt.Println("\tlexed date:", string(date))

	l.consumeSpace()
	next := l.next()
	if next == '!' {
		fmt.Println("\tlexed state:", "!")
		l.consumeSpace()
	} else if next == '*' {
		fmt.Println("\tlexed state:", "*")
		l.consumeSpace()
	} else {
		l.backup()
	}

	payee := l.takeToNextLineOrComment()
	fmt.Println("\tlexed payee:", string(payee))

}

func (l *Lexer) lexPostingLine() {
	l.consumeSpace()

	firstRune := l.next()

	if isCommentIndicator(firstRune) {
		comment := l.takeToNextLine()
		fmt.Println("\tlexed comment:", string(comment))
		return
	}

	if unicode.IsLetter(firstRune) {
		// We need to backup otherwise we'll miss the first rune of the account
		l.backup()
		account := l.takeUntilMoreThanOneSpace()
		fmt.Println("\tlexed account:", string(account))

		// Bail if there are not enough spaces
		if l.consumeSpace() < 2 {
			if len(l.input)-l.pos > 1 {
				log.Fatalln("Not enough spaces following account", len(l.input)-l.pos)
			}
			return
		}

		// Lex the currency
		currency := l.lexCurrency()
		fmt.Println("\tlexed currency:", string(currency))
		l.consumeSpace()

		// Lex the amount
		amount := l.takeToNextLineOrComment()
		fmt.Println("\tlexed amount:", string(amount))

		return
	}

	// If we didn't lex anything, we should reset the parser
	l.backup()
}

// Takes until a number or a space
func (l *Lexer) lexCurrency() []rune {
	runes := make([]rune, 256)
	for {
		r := l.next()
		if r == EOF {
			return runes
		}

		if unicode.IsNumber(r) {
			l.backup()
			return runes
		}

		if r == ' ' || r == '\t' {
			return runes
		}

		runes = append(runes, r)
	}
}

// Move through the bytes of the input, converting to runes as we go
func (l *Lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return EOF
	}

	runeValue, runeWidth := utf8.DecodeRune(l.input[l.pos:])
	l.width = runeWidth
	l.pos += l.width
	return runeValue
}

// Peek at the next rune
func (l *Lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// Steps back one rune
func (l *Lexer) backup() {
	l.pos -= l.width
}

// Consumes spaces
// Returns how many spaces were consumed
func (l *Lexer) consumeSpace() int {
	count := 0
	for {
		r := l.next()
		if r == EOF || !unicode.IsSpace(r) {
			l.backup()
			return count
		}
		if r == '\t' {
			count += TabWidth
		}
		if r == ' ' {
			count++
		}
	}
}

func isCommentIndicator(r rune) bool {
	return r == ';'
}

func countSpace(r rune) int {
	if r == ' ' {
		return 1
	} else if r == '\t' {
		return TabWidth
	}

	return 0
}

func (l *Lexer) takeToNextLine() []rune {
	runes := make([]rune, 256)
	for {
		r := l.next()
		if r == EOF {
			return runes
		}
		runes = append(runes, r)
	}
}

func (l *Lexer) takeToNextLineOrComment() []rune {
	runes := make([]rune, 256)
	for {
		r := l.next()
		if r == EOF {
			return runes
		}

		if isCommentIndicator(r) {
			return runes
		}
		runes = append(runes, r)
	}
}

func (l *Lexer) takeUntilSpace() []rune {
	defer l.backup()
	runes := make([]rune, 256)
	for {
		r := l.next()
		if r == ' ' {
			return runes
		}

		runes = append(runes, r)
	}
}

func (l *Lexer) takeUntilMoreThanOneSpace() []rune {
	// TODO: make this a buffer on the lexer
	runes := make([]rune, 256)
	var previous rune = -1
	for {
		r := l.next()
		if r == EOF {
			return runes
		}

		if r == '\t' {
			return runes
		}

		if r == ' ' {
			if previous != -1 && previous == ' ' {
				return runes
			}
			previous = r
		}
		runes = append(runes, r)
	}
}
