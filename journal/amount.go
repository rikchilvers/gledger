package journal

import (
	"fmt"
)

// Amount encapsulates the quantity of a specific commodity (e.g. GBP)
type Amount struct {
	Commodity string
	Quantity  int64
}

// NewAmount creates an Amount
func NewAmount(q int64) *Amount {
	return &Amount{
		Commodity: "",
		Quantity:  q,
	}
}

func (a Amount) displayableQuantity(withCommodity bool) string {
	q := float64(a.Quantity) / 100
	amount := fmt.Sprintf("%.2f", q)
	if withCommodity {
		return fmt.Sprintf("%s%s", a.Commodity, amount)
	}
	return amount
}
