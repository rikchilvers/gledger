package main

// The date format journal files use
const DateFormat string = "2006-01-02"

// Max line length in runes
const MaxLineLength int = 255

type ParserState int

// The state of the parser (normally, what it is expecting next)
const (
	// Whitespace, comment or date
	AwaitingTransaction ParserState = iota
	TransactionPosting
	Stop
)

func isComment(r rune) bool {
	return r == semicolon || r == hash
}

func isNumeric(r rune) bool {
	return r >= zero && r <= nine
}

func isNewline(r rune) bool {
	return r == newline || r == carriageReturn
}

func isState(r rune) bool {
	return r == exclamation || r == star
}

func toState(r rune) TransactionState {
	switch r {
	case exclamation:
		return Uncleared
	case star:
		return Cleared
	default:
		return NoState
	}
}
