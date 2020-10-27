package journal

import "fmt"

type Posting struct {
	Transaction *Transaction // The transaction this posting belongs to
	Comments    []string     // Any comments attached to the posting
	Account     *Account     // The account this posting relates to
	AccountPath []string     // The : delimited path to the above account
	Amount      *Amount
}

func NewPosting() *Posting {
	return &Posting{
		// id:            uuid.Nil,
		Transaction: nil,
		Comments:    make([]string, 0),
		Account:     nil,
		AccountPath: make([]string, 5),
		Amount:      nil,
	}
}

func (p Posting) String() string {
	return fmt.Sprintf("%s  %s", p.Account.Name, p.Amount.displayableQuantity(true))
}
