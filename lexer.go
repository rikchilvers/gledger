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
	if unicode.IsNumber(firstRune) {
		fmt.Println(("\ttransaction header"))
		return
	}

	// Detect posting lines
	if firstRune == '\t' || firstRune == ' ' && secondRune == ' ' {
		fmt.Println("\tposting line")
		return
	}

	return
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
