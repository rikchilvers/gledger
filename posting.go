package main

type Posting struct {
	comments []string
	account  string
	currency interface{}
	// TODO: represent as a struct with two int fields
	amount interface{}
}

func newPosting(comments []string, account string, currency interface{}, amount interface{}) Posting {
	return Posting{
		comments,
		account,
		currency,
		amount,
	}
}
