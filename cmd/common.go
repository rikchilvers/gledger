package cmd

import (
	"errors"

	"github.com/rikchilvers/gledger/parser"
)

func parse(cmd parser.TransactionHandler) error {
	if rootJournalPath == "" {
		// TODO: use viper to read env variable
		return errors.New("No root journal path provided")
	}

	p := parser.NewParser(cmd)
	if err := p.Parse(rootJournalPath); err != nil {
		return err
	}
	return nil
}
