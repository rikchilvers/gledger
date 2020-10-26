// The lexer reads a file of bytes and identifies components to be parsed

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"unicode"
	"unicode/utf8"
)

//go:generate stringer -type=itemType
type itemType int

const (
	tEmptyLine itemType = iota
	tDate
	tState
	tPayee
	tAccount
	tCommodity
	tAmount
	tComment
	tEOF
)

const eof = -1
const tabWidth = 2 // size of a tab in spaces
const runeBufferCapacity = 256

type lexer struct {
	reader      *bufio.Reader
	input       []byte
	currentLine int
	pos         int // input position
	start       int // item start position
	width       int // width of last element
	parser      journalParser
}

// Lexes the file line by line
func (l *lexer) lex(r io.Reader) error {
	l.reader = bufio.NewReader(r)

	l.currentLine = 1
	for {
		line, isPrefix, err := l.reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
		if isPrefix {
			return errors.New("Unhandled split line (r- todo)")
		}

		// Reset the positions
		l.pos = 0
		l.start = 0
		l.width = 0

		l.input = line
		err = l.lexLine()
		if err != nil {
			return err
		}
		l.currentLine++
	}

	return nil
}

// Lex the line
func (l *lexer) lexLine() error {
	// Bail early if the line is empty
	if len(l.input) == 0 {
		return l.parser.parseItem(tEmptyLine, nil, l.currentLine)
	}

	firstRune := l.next()
	if firstRune == eof {
		return l.parser.parseItem(tEOF, nil, l.currentLine)
	}

	// Detect transaction headers
	// TODO: handle passing different transaction header indicators
	if unicode.IsNumber(firstRune) {
		// Need to backup to include the first rune
		l.backup()
		return l.lexTransactionHeader()
	}

	// Detect posting lines
	if firstRune == '\t' || (firstRune == ' ' && l.peek() == ' ') {
		return l.lexPosting()
	}

	return nil
}

func (l *lexer) lexTransactionHeader() error {
	date := l.takeUntilSpace()
	err := l.parser.parseItem(tDate, date, l.currentLine)
	if err != nil {
		return err
	}

	l.consumeSpace()
	next := l.next()
	if next == '!' {
		l.parser.parseItem(tState, []rune{'!'}, l.currentLine)
		l.consumeSpace()
	} else if next == '*' {
		l.parser.parseItem(tState, []rune{'*'}, l.currentLine)
		l.consumeSpace()
	} else {
		l.backup()
	}

	payee := l.takeToNextLineOrComment()
	err = l.parser.parseItem(tPayee, payee, l.currentLine)
	if err != nil {
		return fmt.Errorf("failed to lex transaction header: %w", err)
	}

	return nil
}

func (l *lexer) lexPosting() error {
	l.consumeSpace()

	firstRune := l.next()

	// TODO: handle comments in postings
	if isCommentIndicator(firstRune) {
		// comment := l.takeToNextLine()
		// l.parser.parseItem(tComment, comment)
		return nil
	}

	if unicode.IsLetter(firstRune) {
		// We need to backup otherwise we'll miss the first rune of the account
		l.backup()
		account := l.takeUntilMoreThanOneSpace()
		err := l.parser.parseItem(tAccount, account, l.currentLine)
		if err != nil {
			return err
		}

		// Bail if there are not enough spaces
		if l.consumeSpace() < 2 {
			if len(l.input)-l.pos > 1 {
				return fmt.Errorf("Not enough spaces following account on line", len(l.input)-l.pos, l.currentLine)
			}
			return nil
		}

		// Lex the commodity
		commodity := l.lexCommodity()
		if l.consumeSpace() > 0 {
			commodity = append(commodity, ' ')
		}
		err = l.parser.parseItem(tCommodity, commodity, l.currentLine)
		if err != nil {
			return err
		}

		// Lex the amount
		amount := l.takeToNextLineOrComment()
		err = l.parser.parseItem(tAmount, amount, l.currentLine)
		if err != nil {
			return err
		}

		return nil
	}

	// If we didn't lex anything, we should reset the parser
	l.backup()

	return nil
}

// Takes until a number or a space
func (l *lexer) lexCommodity() []rune {
	runes := make([]rune, 0, runeBufferCapacity)
	for {
		r := l.next()
		if r == eof {
			return runes
		}

		if unicode.IsNumber(r) || r == '-' || r == '+' {
			l.backup()
			return runes
		}

		if r == ' ' || r == '\t' {
			l.backup()
			return runes
		}

		runes = append(runes, r)
	}
}

// Move through the bytes of the input, converting to runes as we go
func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}

	runeValue, runeWidth := utf8.DecodeRune(l.input[l.pos:])
	l.width = runeWidth
	l.pos += l.width
	return runeValue
}

// Peek at the next rune
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// Steps back one rune
func (l *lexer) backup() {
	l.pos -= l.width
}

// Consumes spaces
// Returns how many spaces were consumed
func (l *lexer) consumeSpace() int {
	count := 0
	for {
		r := l.next()
		if r == eof || !unicode.IsSpace(r) {
			l.backup()
			return count
		}
		if r == '\t' {
			count += tabWidth
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
		return tabWidth
	}

	return 0
}

func (l *lexer) takeToNextLine() []rune {
	runes := make([]rune, 0, runeBufferCapacity)
	for {
		r := l.next()
		if r == eof {
			return runes
		}
		runes = append(runes, r)
	}
}

func (l *lexer) takeToNextLineOrComment() []rune {
	runes := make([]rune, 0, runeBufferCapacity)
	for {
		r := l.next()
		if r == eof {
			return runes
		}

		if isCommentIndicator(r) {
			return runes
		}
		runes = append(runes, r)
	}
}

func (l *lexer) takeUntilSpace() []rune {
	defer l.backup()
	runes := make([]rune, 0, runeBufferCapacity)
	for {
		r := l.next()

		if r == eof {
			return runes
		}

		if r == ' ' {
			return runes
		}

		runes = append(runes, r)
	}
}

func (l *lexer) takeUntilMoreThanOneSpace() []rune {
	// TODO: make this runes slice a buffer on the lexer
	runes := make([]rune, 0, runeBufferCapacity)
	var previous rune = -1
	for {
		r := l.peek()
		if r == eof {
			return runes
		}

		if r == '\t' {
			return runes
		} else if r == ' ' {
			if previous == ' ' {
				l.backup()
				return runes
			}
			previous = r
		} else {
			previous = -1
		}

		l.next()
		runes = append(runes, r)
	}
}

func (l *lexer) takeUntilDecimal() []rune {
	// Take until a decimal or the end of the line
	runes := make([]rune, 0, runeBufferCapacity)
	for {
		r := l.next()
		if r == '.' || r == eof {
			break
		}

		runes = append(runes, r)
	}

	return runes
}
