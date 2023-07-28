package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
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
		err := cli.show(arguments)
		if err != nil {
			log.Fatalf("Cannot show. Reason: %s\n", err)
		} else {
			os.Exit(0)
		}
	}
	if *command == "del" {
		err := cli.del(arguments)
		if err != nil {
			log.Fatalf("Cannot delete. Reason: %s\n", err)
		}
		os.Exit(0)
	}
	if *command == "snooze" {
		err := cli.snooze(arguments)
		if err != nil {
			log.Fatalf("Cannot snooze. Reason: %s\n", err)
		}
		os.Exit(0)
	}
}

func (cli *cli) add(arguments []string) {
	buffer := &bytes.Buffer{}
	for i := 0; i < len(arguments); i++ {
		argument := arguments[i]
		buffer.WriteString(argument)
		if i < len(arguments)-1 {
			buffer.WriteRune(' ')
		}
	}
	title := buffer.String()
	err := cli.app.add(title)
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

func (cli *cli) show(arguments []string) error {
	searchFor := ""
	findAny := false
	if len(arguments) == 0 {
		findAny = true
	} else {
		searchFor = arguments[0]
	}
	if len(arguments) > 1 {
		return errors.New("invalid parameter for show")
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
	return nil
}

func (cli *cli) del(arguments []string) error {
	var searchFor string
	if len(arguments) == 0 {
		return errors.New("invalid parameter for del")
	} else {
		searchFor = arguments[0]
	}
	if len(arguments) > 1 {
		return errors.New("invalid parameter for del")
	}

	entry := cli.app.find(searchFor)

	if entry == nil {
		fmt.Printf("No entry found matching %s\n", searchFor)
		return nil
	}

	return cli.app.delete(entry.Id)
}

func (cli *cli) snooze(arguments []string) error {
	var searchFor string
	snoozeFor := 1 * time.Hour
	if len(arguments) == 0 {
		return errors.New("invalid parameter for snooze")
	} else {
		searchFor = arguments[0]
	}
	if len(arguments) >= 2 {
		var err error
		snoozeFor, err = time.ParseDuration(arguments[1])
		if err != nil {
			return errors.New(fmt.Sprintf("invalid parameter for snooze: %s", err))
		}
	}
	if len(arguments) > 2 {
		return errors.New("invalid parameter for snooze")
	}

	entry := cli.app.find(searchFor)

	if entry == nil {
		fmt.Printf("No entry found matching %s\n", searchFor)
		return nil
	}

	cli.app.setNewDue(entry.Id, time.Now().Add(snoozeFor))
	return nil
}
