package cmd

import (
	"fmt"
	"math"
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
	Short:        "Shows statistics about the journal",
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

const dateLayout string = "2006-01-02"

type journalStatistics struct {
	firstTransactionDate time.Time
	lastTransactionDate  time.Time
	transactionCount     int
	uniqueAccounts       map[string]bool
	uniquePayees         int
}

func newJournalStatistics() journalStatistics {
	return journalStatistics{
		firstTransactionDate: time.Time{},
		lastTransactionDate:  time.Time{},
		transactionCount:     0,
		uniqueAccounts:       make(map[string]bool),
		uniquePayees:         0,
	}
}

func (js *journalStatistics) analyseTransaction(t *journal.Transaction) error {
	// Increment transaction count
	js.transactionCount++

	// Check start date
	if js.firstTransactionDate.IsZero() || t.Date.Before(js.firstTransactionDate) {
		js.firstTransactionDate = t.Date
	}

	// Check end date
	if js.lastTransactionDate.IsZero() || t.Date.After(js.lastTransactionDate) {
		js.lastTransactionDate = t.Date
	}

	for _, p := range t.Postings {
		// Add the account path
		// We don't need to check if it exists beforehand because we don't care about the value
		js.uniqueAccounts[p.AccountPath] = true
	}

	return nil
}

func (js journalStatistics) report() {
	// Report start and end dates
	fmt.Printf("First transaction:\t%s\n", js.firstTransactionDate.Format(dateLayout))
	fmt.Printf("Last transaction:\t%s\n", js.lastTransactionDate.Format(dateLayout))

	// Report duration
	duration := js.lastTransactionDate.Sub(js.firstTransactionDate)
	days := math.Round(duration.Hours() / 24)
	fmt.Printf("Time period:\t\t%.f days\n", days)

	// Report transaction count
	transactionsPerDay := float64(js.transactionCount) / days
	fmt.Printf("Transactions:\t\t%d (%.1f per day)\n", js.transactionCount, transactionsPerDay)

	// Report unique account count
	fmt.Printf("Unique accounts:\t%d\n", len(js.uniqueAccounts))
}
