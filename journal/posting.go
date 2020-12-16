package journal

import "fmt"

// Posting holds details about a single Posting
type Posting struct {
	Transaction *Transaction // The transaction this posting belongs to
	Comments    []string     // Any comments attached to the posting
	Account     *Account     // The account this posting relates to. Set when the parent transaction is linked.
	AccountPath string       // The : delimited path to the above account
	Amount      *Amount
}

// NewPosting creates a Posting
func NewPosting() *Posting {
	return &Posting{
		Transaction: nil,
		Comments:    make([]string, 0),
		Account:     nil,
		AccountPath: "",
		Amount:      nil,
	}
}

func (p *Posting) String() string {
	rs := fmt.Sprintf("%s    %s", p.AccountPath, p.Amount.DisplayableQuantity(true))
	for _, c := range p.Comments {
		rs = fmt.Sprintf("%s\n      ; %s", rs, c)
	}

	return rs
}

// AddComment adds a comment to the posting
func (p *Posting) AddComment(c string) {
	p.Comments = append(p.Comments, c)
}
