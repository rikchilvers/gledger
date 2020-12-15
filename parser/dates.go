package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func parseSmartDate(date string) (time.Time, error) {
	/* Mon Jan 2 15:04:05 MST 2006
	04 - 2006
	07 - 2006/01
	05 - 06/01
	08 - 06/01/02
	*/

	components := strings.Split(date, "/")

	switch len(components) {
	case 1: // year
		return parseYear(components[0])
	case 2: // year/month or month/day
		if len(components[0]) == 4 {
			return parseYearMonth(components)
		}
		return parseMonthDay(components)
	case 3: // year/month/day
		return parseYearMonthDay(components)
	default:
		return time.Time{}, errors.New("unhandled number of smart date components")
	}
}

func parseYear(date string) (time.Time, error) {
	if len(date) != 4 {
		return time.Time{}, fmt.Errorf("could not parse year: '%s'", date)
	}

	year, err := strconv.Atoi(date)
	if err != nil {
		return time.Time{}, err
	}

	return time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local), nil
}

func parseYearMonth(date []string) (time.Time, error) {
	fmt.Println("year/month")

	year := date[0]
	month := date[1]

	switch len(year) + len(month) {
	case 2: // 6/1
	}

	return time.Time{}, nil
}

func parseMonthDay(date []string) (time.Time, error) {
	fmt.Println("month/day")

	return time.Time{}, nil
}

func parseYearMonthDay(date []string) (time.Time, error) {
	fmt.Println("year/month/day")

	return time.Time{}, nil
}
