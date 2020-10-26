package main

import (
	"fmt"
	"os"
)

func main() {
	// parser := NewParser()
	// parser.Parse("test.journal")

	parser := newParser()

	file, err := os.Open("testdata/test.journal")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	err = parser.parse(file)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(parser.journal.rootAccount)
}
