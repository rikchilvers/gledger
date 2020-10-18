package main

// Max line length in runes
const MaxLineLength int = 255

// Number of spaces a tab equates to
const SpacesPerTab int = 2

type ParserState int

// The state of the parser (normally, what it is expecting next)
const (
	// Whitespace, comment or date
	AwaitingTransaction ParserState = iota
	InTransaction
	Stop
)

// Runes
const (
	zero           = rune('0')
	nine           = rune('9')
	tab            = rune('\t')
	newline        = rune('\n')
	carriageReturn = rune('\r')
	space          = rune(' ') // todo - handle all unicode whitespace
	period         = rune('.')
	semicolon      = rune(';')
	exclamation    = rune('!')
	star           = rune('*')
	forwardSlash   = rune('/')
	hash           = rune('#')
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

func toState(r rune) transactionState {
	switch r {
	case exclamation:
		return tUncleared
	case star:
		return tCleared
	default:
		return tNoState
	}
}
