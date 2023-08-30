package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
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

type cli struct {
	app *todoApp
	output
	timeRenderLayout string
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
	due := time.Now().Add(24 * time.Hour)
	lastTitleArgumentIndex := len(arguments) - 1
	if len(arguments) >= 2 {
		parsedDueIn, err := time.ParseDuration(arguments[len(arguments)-1])
		if err == nil {
			due = time.Now().Add(parsedDueIn)
			lastTitleArgumentIndex--
		}
	}
	buffer := &bytes.Buffer{}
	for i := 0; i <= lastTitleArgumentIndex; i++ {
		argument := arguments[i]
		buffer.WriteString(argument)
		if i < len(arguments)-1 {
			buffer.WriteRune(' ')
		}
	}
	title := buffer.String()
	err := cli.app.add(title, due)
	if err != nil {
		cli.Errorf("Could not create %s. Maybe this entry already exists?\n", title)
	}
}

func (cli *cli) list() {
	entries, idMap := cli.app.findAll()

	for _, entry := range entries {
		bold := color.New(color.Bold, color.FgHiBlack).SprintFunc()
		blue := color.New(color.FgBlue).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		dueFunc := green
		if entry.Due.Before(time.Now()) {
			dueFunc = red
		}
		cli.Resultf("[%s] %s %s\n", blue(idMap[entry.Id.String()]), bold(entry.Title), dueFunc(cli.formatRelativeTo(entry.Due, time.Now())))
	}
}

func (cli *cli) due() {
	entries, idMap := cli.app.findWhereDueBefore(time.Now())

	for _, entry := range entries {
		bold := color.New(color.Bold, color.FgHiBlack).SprintFunc()
		blue := color.New(color.FgBlue).SprintFunc()
		cli.Resultf("[%s] %s %s\n", blue(idMap[entry.Id.String()]), bold(entry.Title), cli.formatRelativeTo(entry.Due, time.Now()))
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
		bold := color.New(color.Bold, color.FgHiBlack).SprintFunc()
		blue := color.New(color.FgBlue).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()
		dueFunc := green
		if entry.Due.Before(time.Now()) {
			dueFunc = red
		}
		cli.Resultf("[%s]\n%s\n%s\n%s\n", blue(entryId), bold(entry.Title), dueFunc(cli.format(entry.Due)), entry.Details)
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
	snoozeFor := 1 * time.Hour
	lastTitleArgumentIndex := len(arguments) - 1
	if len(arguments) >= 2 {
		parsedDueIn, err := time.ParseDuration(arguments[len(arguments)-1])
		if err == nil {
			snoozeFor = parsedDueIn
			lastTitleArgumentIndex--
		}
	}
	searchFor := ""
	buffer := &bytes.Buffer{}
	for i := 0; i <= lastTitleArgumentIndex; i++ {
		argument := arguments[i]
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
		err := cli.app.setNewDue(entry.Id, time.Now().Add(snoozeFor))
		if err != nil {
			cli.Errorf("Could not snooze %s %s: %s", entry.Id, entry.Title, err)
		} else {
			cli.Resultf("Snoozed %s %s\n", entry.Id, entry.Title)
		}
	}
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
			if relativeTimestamp.Year() == timestamp.Year() && relativeTimestamp.Month() == timestamp.Month() && relativeTimestamp.Day() == timestamp.Day()-1 {
				return "tomorrow at " + timestamp.Format("15:04")
			}
		}
	} else {

	}
	return cli.format(timestamp)
}
