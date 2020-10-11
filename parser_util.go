package main

// The date format journal files use
const DateFormat string = "2006-01-02"

type ParserState int

// The state of the parser (normally, what it is expecting next)
const (
	// Whitespace, comment or date
	AwaitingTransaction ParserState = iota
	Stop
)
