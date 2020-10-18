package main

import (
	"log"
	"os"
)

func main() {
	// parser := NewParser()
	// parser.Parse("test.journal")

	parser := newParser()

	file, err := os.Open("test.journal")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	parser.parse(file)
}
