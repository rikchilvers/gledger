package cmd

import (
	"fmt"
	"os"

	"github.com/rikchilvers/gledger/parser"
	"github.com/spf13/cobra"
)

var rootJournalPath string

var rootCmd = &cobra.Command{
	Use:   "gledger",
	Short: "gledger - command line budgeting",
	Long:  "gledger is a reimplementation of Ledger in Go\nwith YNAB-style budgeting at its core",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&rootJournalPath, "file", "f", "", "journal file to read (defaults to $LEDGER_FILE)")
}

// Execute runs gledger
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parse(journalPath string) error {
	p := parser.NewParser()
	if err := p.Parse(journalPath); err != nil {
		return err
	}
	return nil
}
