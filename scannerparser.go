package main

import (
	"bufio"
	"log"
	"os"
)

type ScannerParser struct {
	state              ParserState
	scanner            *bufio.Scanner
	line               int
	column             int
	currentTransaction *Transaction
	transactions       []*Transaction
}

func NewScannerParser() *ScannerParser {
	return &ScannerParser{
		state:              AwaitingTransaction,
		scanner:            nil,
		line:               0,
		column:             0,
		currentTransaction: nil,
		transactions:       []*Transaction{},
	}
}

func (p *ScannerParser) advance() {

}

func (p *ScannerParser) Parse(ledgerFile string) {
	file, err := os.Open(ledgerFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	p.scanner = bufio.NewScanner(file)
}
