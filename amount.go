package main

import (
	"fmt"
)

type amount struct {
	commodity string
	quantity  int64
}

func newAmount(q int64) *amount {
	return &amount{
		commodity: "",
		quantity:  q,
	}
}

func (a amount) displayableQuantity(withCommodity bool) string {
	q := float64(a.quantity) / 100
	amount := fmt.Sprintf("%.2f", q)
	if withCommodity {
		return fmt.Sprintf("%s%s", a.commodity, amount)
	}
	return amount
}
