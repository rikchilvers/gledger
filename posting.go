package main

type posting struct {
	comments []string
	account  string
	currency interface{}
	// TODO: represent as a struct with two int fields
	amount interface{}
}

func newPosting(comments []string, account string, currency interface{}, amount interface{}) *posting {
	return &posting{
		comments,
		account,
		currency,
		amount,
	}
}
