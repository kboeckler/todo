package main

import (
	"bytes"
	"strings"
	"time"
)

type Timer struct {
	calculationFunc MapTime
}

type MapTime func(time.Time) time.Time

func (t *Timer) isEmpty() bool {
	return t.calculationFunc == nil
}

func (t *Timer) Resolve(from time.Time) time.Time {
	return t.calculationFunc(from)
}

func ParseTimer(arguments []string, location *time.Location) (string, *Timer) {
	var calculationFunc MapTime = nil
	titleArgs := arguments
	if len(arguments) >= 2 {
		// 1. Is arg duration?
		parsedDueIn, err := time.ParseDuration(arguments[len(arguments)-1])
		if err == nil {
			calculationFunc = inDuration(parsedDueIn)
			if strings.EqualFold("IN", strings.ToUpper(arguments[len(arguments)-2])) {
				titleArgs = arguments[:len(arguments)-2]
			} else {
				titleArgs = arguments[:len(arguments)-1]
			}
		} else {
			// 2. Is Arg tomorrow?
			if strings.EqualFold("TOMORROW", strings.ToUpper(arguments[len(arguments)-1])) {
				calculationFunc = tomorrow(location)
				titleArgs = arguments[:len(arguments)-1]
			} else {
				// 3. Is arg time?
				parsedTime, err := time.ParseInLocation("2006-01-02 15:04", arguments[len(arguments)-1], location)
				if err == nil {
					calculationFunc = atTime(parsedTime)
					if strings.EqualFold("AT", strings.ToUpper(arguments[len(arguments)-2])) {
						titleArgs = arguments[:len(arguments)-2]
					} else {
						titleArgs = arguments[:len(arguments)-1]
					}
				} else if len(arguments) >= 3 {
					// 4. Is arg time with spaces?
					parsedTime, err = time.ParseInLocation("2006-01-02 15:04", arguments[len(arguments)-2]+" "+arguments[len(arguments)-1], location)
					if err == nil {
						calculationFunc = atTime(parsedTime)
						if strings.EqualFold("AT", strings.ToUpper(arguments[len(arguments)-3])) {
							titleArgs = arguments[:len(arguments)-3]
						} else {
							titleArgs = arguments[:len(arguments)-2]
						}
					}
				}
			}
		}
	}
	withoutDuration := ""
	buffer := &bytes.Buffer{}
	for i := 0; i < len(titleArgs); i++ {
		argument := titleArgs[i]
		buffer.WriteString(argument)
		if i < len(titleArgs)-1 {
			buffer.WriteRune(' ')
		}
	}
	withoutDuration = buffer.String()
	return withoutDuration, &Timer{calculationFunc: calculationFunc}
}

func tomorrow(location *time.Location) MapTime {
	return func(relativeTo time.Time) time.Time {
		todayAtZero := time.Date(relativeTo.Year(), relativeTo.Month(), relativeTo.Day(), 0, 0, 0, 0, location)
		tomorrowAtZero := todayAtZero.Add(24 * time.Hour)
		tomorrowAtEleven := tomorrowAtZero.Add(11 * time.Hour)
		return tomorrowAtEleven
	}
}

func atTime(at time.Time) MapTime {
	return func(relativeTo time.Time) time.Time {
		return at
	}
}

func inDuration(in time.Duration) MapTime {
	return func(relativeTo time.Time) time.Time {
		return relativeTo.Add(in)
	}
}
