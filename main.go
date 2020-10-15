package main

import (
	"log"
	"os"
)

func main() {
	// parser := NewParser()
	// parser.Parse("test.journal")

	file, err := os.Open("test.journal")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	lexer := lexer{}
	lexer.lex(file)
}
