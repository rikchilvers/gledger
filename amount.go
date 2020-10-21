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
	offset := 2
	if a.quantity < 0 {
		offset = 3
	}
	amount := fmt.Sprintf("%s.%s", q[:offset], q[len(q)-2:])
	if withCommodity {
		amount = fmt.Sprintf("%s%s", a.commodity, amount)
	}
	return amount
}
