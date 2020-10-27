package main

import (
	"fmt"

	"github.com/rikchilvers/gledger/cmd"
)

func main() {
	cmd.Execute()

	return

	parser := newParser()
	err := parser.parse("testdata/test.journal")
	if err != nil {
		fmt.Println(err)
		return
	}

	printAccountsAndQuantities(*parser.journal.rootAccount, -1)

	// for _, t := range parser.journal.transactions {
	// 	fmt.Println()
	// 	fmt.Println(t)
	// }
}

func printAccountsAndQuantities(a account, depth int) {
	if a.name == "root" {
		for _, child := range a.children {
			printAccountsAndQuantities(*child, depth+1)
		}
		return
	}

	spaces := ""
	for i := 0; i < depth*tabWidth; i++ {
		spaces = fmt.Sprintf(" %s", spaces)
	}
	nameAndQuantity := fmt.Sprintf("\t%s\t%s%s", a.amount.displayableQuantity(true), spaces, a.name)
	fmt.Println(nameAndQuantity)

	for _, child := range a.children {
		printAccountsAndQuantities(*child, depth+1)
	}
}
