package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
	"time"
)

type output struct {
	stdout io.Writer
	stderr io.Writer
}

func (o *output) Resultf(format string, a ...any) {
	_, _ = fmt.Fprintf(o.stdout, format, a...)
}

func (o *output) Errorf(format string, a ...any) {
	_, _ = fmt.Fprintf(o.stderr, format, a...)
}

type MapTime func(time.Time) time.Time

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

type timuration struct {
	calculationFunc MapTime
}

func (t timuration) isEmpty() bool {
	return t.calculationFunc == nil
}

func (t timuration) CalculateFrom(relativeTo time.Time) time.Time {
	return t.calculationFunc(relativeTo)
}

type cli struct {
	app *todoApp
	output
	timeRenderLayout string
	location         *time.Location
}

func (cli *cli) run(args []string) {
	log.Debugf("Running cli with arguments %v", args)
	var command *string
	var arguments []string
	if len(args) > 0 {
		command = &args[0]
		arguments = args[1:]
	} else {
		command = new(string)
	}
	switch *command {
	case "help":
		usage()
	case "add":
		cli.add(arguments)
	case "list":
		cli.list()
	case "due":
		cli.due()
	case "show":
		cli.show(arguments)
	case "del":
		cli.del(arguments)
	case "resolve":
		cli.resolve(arguments)
	case "snooze":
		cli.snooze(arguments)
	default:
		cli.Errorf("command unknown: %s\n", *command)
		usage()
		os.Exit(-1)
	}
}

func (cli *cli) add(arguments []string) {
	var due time.Time
	title, passedTimuration := cli.parseTimuration(arguments)
	if !passedTimuration.isEmpty() {
		due = passedTimuration.CalculateFrom(time.Now())
	} else {
		due = time.Now().Add(24 * time.Hour)
	}
	err := cli.app.add(title, due)
	if err != nil {
		cli.Errorf("Could not create %s. Maybe this entry already exists?\n", title)
	}
}

func (cli *cli) list() {
	entries, idMap := cli.app.findAll()

	for _, entry := range sorted(entries) {
		blue := color.New(color.FgBlue).SprintFunc()
		magenta := color.New(color.FgMagenta).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		dueFunc := green
		if entry.Due.Before(time.Now()) {
			dueFunc = magenta
		}
		cli.Resultf("[%s] %s %s\n", blue(idMap[entry.Id.String()]), entry.Title, dueFunc(cli.formatRelativeTo(entry.Due, time.Now())))
	}
}

func (cli *cli) due() {
	entries, idMap := cli.app.findWhereDueBefore(time.Now())

	for _, entry := range sorted(entries) {
		blue := color.New(color.FgBlue).SprintFunc()
		magenta := color.New(color.FgMagenta).SprintFunc()
		cli.Resultf("[%s] %s %s\n", blue(idMap[entry.Id.String()]), entry.Title, magenta(cli.formatRelativeTo(entry.Due, time.Now())))
	}
}

func (cli *cli) show(arguments []string) {
	searchFor := ""
	buffer := &bytes.Buffer{}
	for i, argument := range arguments {
		buffer.WriteString(argument)
		if i < len(arguments)-1 {
			buffer.WriteRune(' ')
		}
	}
	searchFor = buffer.String()

	var entry *todo
	var entryId string

	if len(searchFor) > 0 {
		entry, entryId = cli.app.find(searchFor)
	}

	if entry == nil {
		cli.Errorf("No entry found matching %s\n", searchFor)
	} else {
		blue := color.New(color.FgBlue).SprintFunc()
		magenta := color.New(color.FgMagenta).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		dueFunc := green
		if entry.Due.Before(time.Now()) {
			dueFunc = magenta
		}
		cli.Resultf("[%s]\n%s\n%s\n%s\n", blue(entryId), entry.Title, dueFunc(cli.format(entry.Due)), entry.Details)
	}
}

func (cli *cli) del(arguments []string) {
	searchFor := ""
	buffer := &bytes.Buffer{}
	for i, argument := range arguments {
		buffer.WriteString(argument)
		if i < len(arguments)-1 {
			buffer.WriteRune(' ')
		}
	}
	searchFor = buffer.String()

	var entry *todo

	if len(searchFor) > 0 {
		entry, _ = cli.app.find(searchFor)
	}

	if entry == nil {
		cli.Errorf("No entry found matching %s\n", searchFor)
	} else {
		err := cli.app.delete(entry.Id)
		if err != nil {
			cli.Errorf("Could not delete %s %s: %s", entry.Id, entry.Title, err)
		} else {
			cli.Resultf("Deleted %s %s\n", entry.Id, entry.Title)
		}
	}

}

func (cli *cli) resolve(arguments []string) {
	searchFor := ""
	buffer := &bytes.Buffer{}
	for i, argument := range arguments {
		buffer.WriteString(argument)
		if i < len(arguments)-1 {
			buffer.WriteRune(' ')
		}
	}
	searchFor = buffer.String()

	var entry *todo

	if len(searchFor) > 0 {
		entry, _ = cli.app.find(searchFor)
	}

	if entry == nil {
		cli.Errorf("No entry found matching %s\n", searchFor)
	} else {
		err := cli.app.resolve(entry.Id)
		if err != nil {
			cli.Errorf("Could not resolve %s %s: %s", entry.Id, entry.Title, err)
		} else {
			cli.Resultf("Resolved %s %s\n", entry.Id, entry.Title)
		}
	}
}

func (cli *cli) snooze(arguments []string) {
	var newDue time.Time
	searchFor, passedTimuration := cli.parseTimuration(arguments)
	if !passedTimuration.isEmpty() {
		newDue = passedTimuration.CalculateFrom(time.Now())
	} else {
		newDue = time.Now().Add(1 * time.Hour)
	}

	var entry *todo

	if len(searchFor) > 0 {
		entry, _ = cli.app.find(searchFor)
	}

	if entry == nil {
		cli.Errorf("No entry found matching %s\n", searchFor)
	} else {
		err := cli.app.setNewDue(entry.Id, newDue)
		if err != nil {
			cli.Errorf("Could not snooze %s %s: %s", entry.Id, entry.Title, err)
		} else {
			cli.Resultf("Snoozed %s %s\n", entry.Id, entry.Title)
		}
	}
}

func (cli *cli) parseTimuration(arguments []string) (string, timuration) {
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
				calculationFunc = tomorrow(cli.location)
				titleArgs = arguments[:len(arguments)-1]
			} else {
				// 3. Is arg time?
				parsedTime, err := time.ParseInLocation("2006-01-02 15:04", arguments[len(arguments)-1], cli.location)
				if err == nil {
					calculationFunc = atTime(parsedTime)
					if strings.EqualFold("AT", strings.ToUpper(arguments[len(arguments)-2])) {
						titleArgs = arguments[:len(arguments)-2]
					} else {
						titleArgs = arguments[:len(arguments)-1]
					}
				} else if len(arguments) >= 3 {
					// 4. Is arg time with spaces?
					parsedTime, err = time.ParseInLocation("2006-01-02 15:04", arguments[len(arguments)-2]+" "+arguments[len(arguments)-1], cli.location)
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
	return withoutDuration, timuration{calculationFunc: calculationFunc}
}

func (cli *cli) format(timestamp time.Time) string {
	return timestamp.Format(cli.timeRenderLayout)
}

func (cli *cli) formatRelativeTo(timestamp, relativeTimestamp time.Time) string {
	dueIn := timestamp.Sub(relativeTimestamp)
	if dueIn >= 0 {
		if dueIn <= 12*time.Hour {
			return "in " + dueIn.String()
		} else {
			relativeTomorrow := relativeTimestamp.Add(24 * time.Hour)
			if relativeTomorrow.Year() == timestamp.Year() && relativeTomorrow.Month() == timestamp.Month() && relativeTomorrow.Day() == timestamp.Day() {
				return timestamp.Format("tomorrow at 15:04")
			}
			relative2d := relativeTimestamp.Add(48 * time.Hour)
			if relative2d.Year() == timestamp.Year() && relative2d.Month() == timestamp.Month() && relative2d.Day() == timestamp.Day() {
				return "in 2 days"
			}
			relative3d := relativeTimestamp.Add(72 * time.Hour)
			if relative3d.Year() == timestamp.Year() && relative3d.Month() == timestamp.Month() && relative3d.Day() == timestamp.Day() {
				return "in 3 days"
			}
			return timestamp.Format("at Mon, 02 Jan 2006")
		}
	} else {
		if dueIn >= -12*time.Hour {
			return "for " + dueIn.Abs().String()
		}
		relativeYesterday := relativeTimestamp.Add(-24 * time.Hour)
		if relativeYesterday.Year() == timestamp.Year() && relativeYesterday.Month() == timestamp.Month() && relativeYesterday.Day() == timestamp.Day() {
			return "since yesterday"
		}
		relative2d := relativeTimestamp.Add(-48 * time.Hour)
		if relative2d.Year() == timestamp.Year() && relative2d.Month() == timestamp.Month() && relative2d.Day() == timestamp.Day() {
			return "for 2 days"
		}
		relative3d := relativeTimestamp.Add(-72 * time.Hour)
		if relative3d.Year() == timestamp.Year() && relative3d.Month() == timestamp.Month() && relative3d.Day() == timestamp.Day() {
			return "for 3 days"
		}
		return timestamp.Format("since Mon, 02 Jan 2006")
	}
}
