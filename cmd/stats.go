package cmd

import (
	"time"

	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:          "stats",
	Aliases:      []string{"statistics", "s"},
	Short:        "Shows some journal statistics",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		js := newJournalStatistics()
		if err := parse(js.analyseTransaction); err != nil {
			return err
		}
		js.report()
		return nil
	},
}

type journalStatistics struct {
	firstTransaction time.Time
	lastTransaction  time.Time
	uniqueAccounts   int
	uniquePayees     int
}

func newJournalStatistics() journalStatistics {
	return journalStatistics{
		firstTransaction: time.Time{},
		lastTransaction:  time.Time{},
		uniqueAccounts:   0,
		uniquePayees:     0,
	}
}

func (js *journalStatistics) analyseTransaction(t *journal.Transaction) error {
	return nil
}

func (js journalStatistics) report() {

}
