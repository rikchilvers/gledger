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

	var components []string
	var separator rune

	if strings.IndexRune(date, '/') > 0 {
		components = strings.Split(date, "/")
		separator = '/'
	} else if strings.IndexRune(date, '.') > 0 {
		components = strings.Split(date, ".")
		separator = '.'
	} else if strings.IndexRune(date, '-') > 0 {
		components = strings.Split(date, "-")
		separator = '-'
	} else {
		components = append(components, date)
	}

	switch len(components) {
	case 1: // year
		return parseYear(date)
	case 2: // year/month or month/day
		if len(components[0]) == 4 {
			return parseYearMonth(date, separator)
		}
		return parseMonthDay(date, separator)
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

// parseYearMonth expects dates in the format 2020/06
func parseYearMonth(date string, separator rune) (time.Time, error) {
	var format string
	switch separator {
	case '/':
		format = "2006/01"
	case '.':
		format = "2006.01"
	case '-':
		format = "2006-01"
	default:
		return time.Time{}, errors.New("unhandled date format")
	}

	parsed, err := time.ParseInLocation(format, date, time.Local)
	if err != nil {
		return time.Time{}, err
	}

	return parsed, nil
}

// parseMonthDaty expected dates in the format 06/22
func parseMonthDay(date string, separator rune) (time.Time, error) {
	year := time.Now().Year()
	var format, dateWithYear string
	switch separator {
	case '/':
		format = "2006/01/02"
		dateWithYear = fmt.Sprintf("%d/%s", year, date)
	case '.':
		format = "2006.01.02"
		dateWithYear = fmt.Sprintf("%d.%s", year, date)
	case '-':
		format = "2006-01-02"
		dateWithYear = fmt.Sprintf("%d-%s", year, date)
	default:
		return time.Time{}, errors.New("unhandled date format")
	}

	parsed, err := time.ParseInLocation(format, dateWithYear, time.Local)
	if err != nil {
		return time.Time{}, err
	}

	return parsed, nil
}

// parseYearMonthDay expects dates in the format 2020/06/22
func parseYearMonthDay(date []string) (time.Time, error) {
	fmt.Println("year/month/day")

	return time.Time{}, nil
}
