package main

type amount struct {
	commodity string
	quantity  int64
}

// TODO: during parsing, check amounts with commodities cannot be created without amounts

func newAmount() *amount {
	return &amount{
		commodity: "",
		quantity:  0,
	}
}
