package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"time"
)

type cli struct {
	app *todoApp
}

func (cli *cli) run(args []string) {
	var command *string
	var arguments []string
	if len(args) > 0 {
		command = &args[0]
		arguments = args[1:]
	}
	if command == nil {
		// default case print usage
		flag.Usage()
		os.Exit(-1)
	}
	if *command == "help" {
		flag.Usage()
		os.Exit(0)
	}
	if *command == "add" {
		cli.add(arguments)
		os.Exit(0)
	}
	if *command == "list" {
		cli.list()
		os.Exit(0)
	}
	if *command == "due" {
		cli.due()
		os.Exit(0)
	}
	if *command == "show" {
		cli.show(arguments)
		os.Exit(0)
	}
	if *command == "del" {
		cli.del(arguments)
		os.Exit(0)
	}
	if *command == "resolve" {
		cli.resolve(arguments)
		os.Exit(0)
	}
	if *command == "snooze" {
		cli.snooze(arguments)
		os.Exit(0)
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
		fmt.Printf("Could not create %s. Maybe this entry already exists?\n", title)
	}
}

func (cli *cli) list() {
	entries := cli.app.findAll()

	for _, entry := range entries {
		fmt.Printf("%s Title: %s, Details: %s, Due: %s, Notification: %v\n", entry.Id, entry.Title, entry.Details, entry.Due, entry.Notification)
	}
}

func (cli *cli) due() {
	entries := cli.app.findWhereDueBefore(time.Now())

	for _, entry := range entries {
		fmt.Printf("%s Title: %s, Details: %s, Due: %s, Notification: %v\n", entry.Id, entry.Title, entry.Details, entry.Due, entry.Notification)
	}
}

func (cli *cli) show(arguments []string) {
	searchFor := ""
	findAny := false
	if len(arguments) == 0 {
		findAny = true
	} else {
		buffer := &bytes.Buffer{}
		for i, argument := range arguments {
			buffer.WriteString(argument)
			if i < len(arguments)-1 {
				buffer.WriteRune(' ')
			}
		}
		searchFor = buffer.String()
	}

	var entry *todo

	if findAny {
		entries := cli.app.findAll()
		if len(entries) > 0 {
			entry = &entries[0]
		}
	} else {
		entry = cli.app.find(searchFor)
	}

	if entry == nil {
		fmt.Printf("No entry found matching %s\n", searchFor)
	} else {
		fmt.Printf("%s Title: %s, Details: %s, Due: %s, Notification: %v\n", entry.Id, entry.Title, entry.Details, entry.Due, entry.Notification)
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
		entry = cli.app.find(searchFor)
	}

	if entry == nil {
		fmt.Printf("No entry found matching %s\n", searchFor)
	} else {
		cli.app.delete(entry.Id)
		fmt.Printf("Deleted %s %s\n", entry.Id, entry.Title)
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
		entry = cli.app.find(searchFor)
	}

	if entry == nil {
		fmt.Printf("No entry found matching %s\n", searchFor)
	} else {
		cli.app.resolve(entry.Id)
		fmt.Printf("Resolved %s %s\n", entry.Id, entry.Title)
	}
}

func (cli *cli) snooze(arguments []string) {
	searchFor := ""
	snoozeFor := 1 * time.Hour
	lastTitleArgumentIndex := len(arguments) - 1
	if len(arguments) >= 2 {
		parsedDueIn, err := time.ParseDuration(arguments[len(arguments)-1])
		if err == nil {
			snoozeFor = parsedDueIn
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
	searchFor = buffer.String()

	var entry *todo

	if len(searchFor) > 0 {
		entry = cli.app.find(searchFor)
	}

	if entry == nil {
		fmt.Printf("No entry found matching %s\n", searchFor)
	} else {
		cli.app.setNewDue(entry.Id, time.Now().Add(snoozeFor))
		fmt.Printf("Snoozed %s %s\n", entry.Id, entry.Title)
	}
}
