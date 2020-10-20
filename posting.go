package main

import "fmt"

type posting struct {
	// id            uuid.UUID // Identifier for the posting
	transaction *transaction // The transaction this posting belongs to
	comments    []string     // Any comments attached to the posting
	account     *account     // The account this posting relates to
	amount      *amount
}

func newPosting() *posting {
	return &posting{
		// id:            uuid.Nil,
		transaction: nil,
		comments:    make([]string, 0),
		account:     nil,
		amount:      nil,
	}
}

func (p posting) String() string {
	return fmt.Sprintf("%s  %s", p.account.name, p.amount.displayableQuantity(true))
}
