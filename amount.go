package main

import (
	"fmt"
	"strconv"
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
	q := strconv.FormatInt(a.quantity, 10)
	amount := fmt.Sprintf("%s.%s", q[:2], q[len(q)-2:])
	// amount := fmt.Sprintf("%d", a.quantity)
	if withCommodity {
		amount = fmt.Sprintf("%s%s", a.commodity, amount)
	}
	return amount
}
