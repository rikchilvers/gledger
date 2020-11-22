// The lexer reads a file of bytes and identifies components to be parsed

package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"unicode"
	"unicode/utf8"

	"github.com/rikchilvers/gledger/shared"
)

//go:generate stringer -type=itemType
type itemType int

const (
	emptyLineItem itemType = iota
	includeItem
	dateItem
	stateItem
	payeeItem
	accountItem
	commodityItem
	amountItem
	commentItem
	eofItem
)

const eof = -1
const runeBufferCapacity = 256

type lexer struct {
	locationHint string
	reader       *bufio.Reader
	input        []byte
	currentLine  int
	pos          int // input position
	start        int // item start position
	width        int // width of last element
	parser       itemParser
}

func newLexer(reader io.Reader, locationHint string, parser itemParser) lexer {
	return lexer{
		locationHint: locationHint,
		reader:       bufio.NewReader(reader),
		input:        nil,
		currentLine:  0,
		pos:          0,
		start:        0,
		width:        0,
		parser:       parser,
	}
}

// Lexes the file line by line
func (l *lexer) lex() error {
	l.currentLine = 1
	for {
		line, isPrefix, err := l.reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				// Let the parser know we have reached the end of the file
				parseError := l.parser(eofItem, nil)
				if parseError != nil {
					return parseError
				}
				break
			}
			return err
		}
		if isPrefix {
			log.Fatalln("Unhandled split line (todo)")
		}

		// Reset the positions
		l.pos = 0
		l.start = 0
		l.width = 0

		l.input = line
		if err = l.lexLine(); err != nil {
			// Add line data and pass up the error
			return fmt.Errorf("%s:%d\n%w", l.locationHint, l.currentLine, err)
		}
		l.currentLine++
	}

	return nil
}

// Lex the line
func (l *lexer) lexLine() error {
	// Bail early if the line is empty
	if len(l.input) == 0 {
		return l.parser(emptyLineItem, nil)
	}

	// We take the first rune rather than peeking so that we can detect posting lines
	firstRune := l.next()

	// Bail if the line is a comment
	if isCommentIndicator(firstRune) {
		return nil
	}

	// Handle include directives
	if firstRune == 'i' {
		l.backup()
		return l.lexIncludeDirective()
	}

	// Handle EOF
	// This will probably only be called during tests
	if firstRune == eof {
		return l.parser(eofItem, nil)
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

	return errors.New("unexpected line type during lexing")
}

func (l *lexer) lexIncludeDirective() error {
	if directive := l.takeUntilSpace(); !equal(directive, []rune("include")) {
		return fmt.Errorf("unexpected directive: %s", string(directive))
	}

	if l.consumeSpace() == 0 {
		return errors.New("could not lex include directive")
	}

	fileToInclude := l.takeToNextLineOrComment()

	if len(fileToInclude) == 0 {
		return errors.New("could not lex include directive")
	}

	// We need to add any directory information here
	// because the parser does not know about where we're including from
	dir, _ := filepath.Split(l.locationHint)
	fileToInclude = append([]rune(dir), fileToInclude...)

	return l.parser(includeItem, fileToInclude)
}

// Equal tells whether a and b contain the same elements.
// A nil argument is equivalent to an empty slice.
func equal(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func (l *lexer) lexTransactionHeader() error {
	var err error
	date := l.takeUntilSpace()

	if err = l.parser(dateItem, date); err != nil {
		return err
	}

	l.consumeSpace()
	next := l.next()
	if next == '!' {
		err = l.parser(stateItem, []rune{'!'})
		l.consumeSpace()
	} else if next == '*' {
		err = l.parser(stateItem, []rune{'*'})
		l.consumeSpace()
	} else {
		l.backup()
	}
	if err != nil {
		return err
	}

	payee := l.takeToNextLineOrComment()
	if err = l.parser(payeeItem, payee); err != nil {
		return err
	}

	return nil
}

func (l *lexer) lexPosting() error {
	var err error
	l.consumeSpace()

	firstRune := l.next()

	// TODO: handle comments in postings
	if isCommentIndicator(firstRune) {
		// comment := l.takeToNextLine()
		// l.parser.parseItem(commentItem, comment)
		return nil
	}

	if unicode.IsLetter(firstRune) {
		// We need to backup otherwise we'll miss the first rune of the account
		l.backup()
		account := l.takeUntilMoreThanOneSpace()
		if err = l.parser(accountItem, account); err != nil {
			return err
		}

		// Bail if there are not enough spaces
		if l.consumeSpace() < 2 {
			if len(l.input)-l.pos > 1 {
				return errors.New("not enough spaces following account")
			}
			return nil
		}

		// Lex the commodity
		commodity := l.lexCommodity()
		if l.consumeSpace() > 0 {
			commodity = append(commodity, ' ')
		}
		if err = l.parser(commodityItem, commodity); err != nil {
			return err
		}

		// Lex the amount
		amount := l.takeToNextLineOrComment()
		if err = l.parser(amountItem, amount); err != nil {
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
			count += shared.TabWidth
		}
		if r == ' ' {
			count++
		}
	}
}

func isCommentIndicator(r rune) bool {
	return r == ';' || r == '#'
}

func countSpace(r rune) int {
	if r == ' ' {
		return 1
	} else if r == '\t' {
		return shared.TabWidth
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

		if r == eof || r == '\t' {
			if previous == ' ' {
				// Drop the trailing space
				return runes[:len(runes)-1]
			}
			return runes
		}

		if r == ' ' {
			if previous == ' ' {
				l.backup()
				// Drop the space we appended last time
				return runes[:len(runes)-1]
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
