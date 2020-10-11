package main

type Posting struct {
	account  string
	currency interface{}
	// TODO: represent as a struct with two int fields
	amount interface{}
}

func newPosting(account string, currency interface{}, amount interface{}) Posting {
	return Posting{
		account,
		currency,
		amount,
	}
}
