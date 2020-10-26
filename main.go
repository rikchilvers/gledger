package main

import (
	"fmt"
)

func main() {
	parser := newParser()
	err := parser.parse("testdata/test.journal")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(parser.journal.rootAccount)
	for _, t := range parser.journal.transactions {
		fmt.Println()
		fmt.Println(t)
	}
}
