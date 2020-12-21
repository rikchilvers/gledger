package cmd

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/rikchilvers/gledger/journal"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(statsCmd)
}

var statsCmd = &cobra.Command{
	Use:          "statistics",
	Aliases:      []string{"stats", "s"},
	Short:        "Shows statistics about the journal",
	SilenceUsage: true,
	Run: func(_ *cobra.Command, _ []string) {
		sj := newStatisticsJournal()
		th := dateCheckedFilteringTransactionHandler(sj.transactionHandler)
		if err := parse(th, nil); err != nil {
			fmt.Println(err)
			return
		}
		sj.prepare()
		sj.report()
	},
}

const dateLayout string = "2006-01-02"

type statisticsJournal struct {
	firstTransactionDate time.Time
	lastTransactionDate  time.Time
	transactionCount     int
	uniqueAccounts       map[string]bool
	uniquePayees         map[string]bool
	journalFiles         map[string]bool
	incomeBuckets        map[time.Time]int64
	expenseBuckets       map[time.Time]int64
	ageOfMoney           float64
}

func newStatisticsJournal() statisticsJournal {
	return statisticsJournal{
		firstTransactionDate: time.Time{},
		lastTransactionDate:  time.Time{},
		transactionCount:     0,
		uniqueAccounts:       make(map[string]bool),
		uniquePayees:         make(map[string]bool),
		journalFiles:         make(map[string]bool),
		incomeBuckets:        make(map[time.Time]int64),
		expenseBuckets:       make(map[time.Time]int64),
		ageOfMoney:           0.0,
	}
}

func (js *statisticsJournal) transactionHandler(t *journal.Transaction, path string) error {
	// Log the file
	js.journalFiles[path] = true

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

	// Add the account path
	for _, p := range t.Postings {
		// We don't need to check if it exists beforehand because we don't care about the value
		js.uniqueAccounts[p.AccountPath] = true
	}

	// Add the payee
	js.uniquePayees[t.Payee] = true

	// Add income and expenses for age of money calculation
	for _, p := range t.Postings {
		components := strings.Split(p.AccountPath, ":")

		if components[0] == journal.IncomeID {
			js.incomeBuckets[t.Date] -= p.Amount.Quantity
		}

		if components[0] == journal.ExpensesID {
			js.expenseBuckets[t.Date] += p.Amount.Quantity
		}
	}

	return nil
}

func (js *statisticsJournal) prepare() {
	// Sort income
	incomeKeys := make([]time.Time, 0, len(js.incomeBuckets))
	for k := range js.incomeBuckets {
		incomeKeys = append(incomeKeys, k)
	}
	sort.Slice(incomeKeys, func(i, j int) bool {
		return incomeKeys[i].Before(incomeKeys[j])
	})

	// Sort expenses
	expenseKeys := make([]time.Time, 0, len(js.expenseBuckets))
	for k := range js.expenseBuckets {
		expenseKeys = append(expenseKeys, k)
	}
	sort.Slice(expenseKeys, func(i, j int) bool {
		return expenseKeys[i].Before(expenseKeys[j])
	})

	// Calculate age of money
	ages := make([]time.Duration, 0)

expenseLoop:
	for _, ek := range expenseKeys {
		expense := js.expenseBuckets[ek]

		for _, ik := range incomeKeys {
			income := js.incomeBuckets[ik]
			if income == 0 {
				continue
			}

			duration := ek.Sub(ik)
			ages = append(ages, duration)

			// Handle not having enough in the income bucket
			if expense > income {
				js.incomeBuckets[ik] = 0
				expense -= income
			} else {
				js.incomeBuckets[ik] = income - expense
				continue expenseLoop
			}
		}
	}

	rangeCap := 0
	if len(ages) > 20 {
		rangeCap = 11
	}
	var summedAges time.Duration = 0
	for _, a := range ages[:len(ages)-rangeCap] { // we only care about the final 10
		summedAges += a
	}

	// filter arguments could mean there were no transactions to process
	// so we need to guard against dividing by 0 later
	count := int64(math.Max(1, float64(len(ages))))
	summedAges = time.Duration(summedAges.Nanoseconds() / count)

	js.ageOfMoney = math.Max(summedAges.Hours()/24, 0)
}

func (js *statisticsJournal) report() {
	if js.transactionCount == 0 {
		fmt.Println("No transactions matched arguments.")
		return
	}

	// Report the files
	fmt.Printf("Transactions found in %d files:\n", len(js.journalFiles))
	for p := range js.journalFiles {
		fmt.Printf("  %s\n", p)
	}

	// Report start and end dates
	fmt.Printf("First transaction:\t%s (%s)\n", js.firstTransactionDate.Format(dateLayout), humanize.Time(js.firstTransactionDate))
	fmt.Printf("Last transaction:\t%s (%s)\n", js.lastTransactionDate.Format(dateLayout), humanize.Time(js.lastTransactionDate))

	// Report duration
	duration := js.lastTransactionDate.Sub(js.firstTransactionDate)
	days := math.Round(duration.Hours() / 24)
	fmt.Printf("Time period:\t\t%.f days\n", days)

	// Report transaction count
	transactionsPerDay := float64(js.transactionCount) / days
	fmt.Printf("Transactions:\t\t%d (%.1f per day)\n", js.transactionCount, transactionsPerDay)

	// Report number of unique accounts
	fmt.Printf("Unique accounts:\t%d\n", len(js.uniqueAccounts))

	// Report number of unique payees
	fmt.Printf("Unique payees:\t\t%d\n", len(js.uniquePayees))

	// Report age of money
	fmt.Printf("Age of money:\t\t%.f days\n", js.ageOfMoney)
}
