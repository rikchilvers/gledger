package main

type Posting struct {
	account  string
	currency interface{}
	amount   interface{}
}

func newPosting(account, currency string, amount interface{}) Posting {
	return Posting{
		account,
		currency,
		amount,
	}
}
