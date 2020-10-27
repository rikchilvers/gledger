package main

import (
	"github.com/rikchilvers/gledger/cmd"
)

func main() {
	cmd.Execute()
}

/*
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
*/
