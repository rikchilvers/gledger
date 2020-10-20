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

//go:generate stringer -type=itemType
type itemType int

const (
	tDate itemType = iota
	tState
	tPayee
	tAccount
	tCommodity
	tAmount
	tComment
)

const eof = -1
const tabWidth = 2 // size of a tab in spaces
const runeBufferCapacity = 256

type lexer struct {
	reader *bufio.Reader
	input  []byte
	pos    int // input position
	start  int // item start position
	width  int // width of last element
	parser *parser
}

// Lexes the file line by line
func (l *lexer) lex(r io.Reader) {
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
func (l *lexer) lexLine() {
	// Bail early if the line is empty
	if len(l.input) == 0 {
		return
	}

	firstRune := l.next()
	if firstRune == eof {
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
		l.lexPosting()
		return
	}

	return
}

func (l *lexer) lexTransactionHeader() {
	// Need to backup to include the first rune
	l.backup()

	date := l.takeUntilSpace()
	fmt.Println("\tlexed date:", string(date))
	l.parser.parseItem(tDate, date)

	l.consumeSpace()
	next := l.next()
	if next == '!' {
		fmt.Println("\tlexed state:", "!")
		l.parser.parseItem(tState, []rune{'!'})
		l.consumeSpace()
	} else if next == '*' {
		fmt.Println("\tlexed state:", "*")
		l.parser.parseItem(tState, []rune{'*'})
		l.consumeSpace()
	} else {
		l.backup()
	}

	payee := l.takeToNextLineOrComment()
	fmt.Println("\tlexed payee:", string(payee))
	l.parser.parseItem(tPayee, payee)
}

func (l *lexer) lexPosting() {
	l.consumeSpace()

	firstRune := l.next()

	if isCommentIndicator(firstRune) {
		comment := l.takeToNextLine()
		fmt.Println("\tlexed comment:", string(comment))
		// l.parser.parseItem(tComment, comment)
		return
	}

	if unicode.IsLetter(firstRune) {
		// We need to backup otherwise we'll miss the first rune of the account
		l.backup()
		account := l.takeUntilMoreThanOneSpace()
		fmt.Println("\tlexed account:", string(account))
		l.parser.parseItem(tAccount, account)

		// Bail if there are not enough spaces
		if l.consumeSpace() < 2 {
			if len(l.input)-l.pos > 1 {
				log.Fatalln("Not enough spaces following account", len(l.input)-l.pos)
			}
			return
		}

		// Lex the commodity
		commodity := l.lexCommodity()
		fmt.Println("\tlexed commodity:", string(commodity))
		if l.consumeSpace() > 0 {
			commodity = append(commodity, ' ')
		}
		l.parser.parseItem(tCommodity, commodity)

		// Lex the amount
		amount := l.takeToNextLineOrComment()
		fmt.Println("\tlexed amount:", string(amount))
		l.parser.parseItem(tAmount, amount)

		return
	}

	// If we didn't lex anything, we should reset the parser
	l.backup()
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
