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
func NewAmount(c string, q int64) *Amount {
	return &Amount{
		Commodity: c,
		Quantity:  q,
	}
}

func (a Amount) String() string {
	return a.DisplayableQuantity(false)
}

// DisplayableQuantity formats the Amount's commodity and quantity
func (a Amount) DisplayableQuantity(withCommodity bool) string {
	q := float64(a.Quantity) / 100
	amount := fmt.Sprintf("%.2f", q)
	if withCommodity {
		return fmt.Sprintf("%s%s", a.Commodity, amount)
	}
	return amount
}

// Add adds an amount to this one
func (a *Amount) Add(other Amount) error {
	if a.Commodity != other.Commodity {
		// return errors.New("unhandled addition of unmatched commodities")
		fmt.Printf("unhandled addition of unmatched commodities: '%s' and '%s'\n", a.Commodity, other.Commodity)
	}
	a.Quantity += other.Quantity
	return nil
}

// Subtract subtracts an amount from this one
func (a *Amount) Subtract(other Amount) error {
	if a.Commodity != other.Commodity {
		// return errors.New("unhandled addition of unmatched commodities")
		fmt.Printf("unhandled subtraction of unmatched commodities: '%s' and '%s'\n", a.Commodity, other.Commodity)
	}
	a.Quantity -= other.Quantity
	return nil
}
