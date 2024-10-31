package main

import (
	"bytes"
	"fmt"
	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"os/exec"
	"regexp"
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

type cli struct {
	app *todoApp
	cfg config
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
	title, timer := ParseTimer(arguments, cli.location)
	if !timer.isEmpty() {
		due = timer.Resolve(time.Now())
	} else {
		due = time.Now().Add(24 * time.Hour)
	}

	userInput := cli.createDescriptionInput(title, due)
	editorUserInput, err := cli.openInEditor(userInput)
	if err != nil {
		log.Debugf("Error processing input in editor: %v", err)
	}
	cleansedUserInput := cli.cleanseInput(editorUserInput)

	if len(cleansedUserInput) == 0 {
		cli.Errorf("Skip creating %s due to an empty title.\n", title)
	} else {
		userTitle, userDescription := cli.parseDescriptionInput(cleansedUserInput)
		err := cli.app.add(userTitle, userDescription, due)
		if err != nil {
			cli.Errorf("Could not create %s. Maybe this entry already exists?\n", userTitle)
		}
	}

}

func (cli *cli) createDescriptionInput(title string, due time.Time) string {
	return fmt.Sprintf(`%s
# Please enter the title of your todo, adding a description after
# an empty line if needed. Lines starting with '#' will be ignored,
# and an empty input aborts this command.
#
# Title from command: %s
# Due date of this todo: %s
`, title, title, cli.format(due))
}

func (cli *cli) parseDescriptionInput(input string) (title string, description string) {
	split := strings.Split(input, "\n\n")
	title = strings.TrimSpace(split[0])
	description = ""
	if len(split) > 1 {
		description = strings.TrimSpace(strings.Replace(input, title+"\n\n", "", 1))
	}
	return
}

func (cli *cli) cleanseInput(input string) string {
	commentRegex, err := regexp.Compile("#.*\\n")
	if err != nil {
		panic(err)
	}
	replacedString := commentRegex.ReplaceAllString(input, "")
	trimmed := strings.Trim(replacedString, " \n")
	multiNewLineRegex, err := regexp.Compile("\\n[ \\n]+\\n")
	if err != nil {
		panic(err)
	}
	return multiNewLineRegex.ReplaceAllString(trimmed, "\n\n")
}

func (cli *cli) openInEditor(input string) (string, error) {
	_, err := newFileWriter("todo.tmp", false).Write([]byte(input))
	if err != nil {
		return input, err
	}
	defer newFileDeleter("todo.tmp").Delete()
	cmd := exec.Command(cli.cfg.EditorCmd, "todo.tmp")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	log.Debugf("Calling editor command: %s", cmd)
	err = cmd.Run()
	if err == nil {
		log.Debugf("Executing editor command was successful")
	} else {
		log.Debugf("Error of executing editor command: %s", err)
		return input, err
	}
	result, err := newFileReader("todo.tmp").ReadString()
	if err != nil {
		return input, err
	}
	return result, nil
}

func (cli *cli) list() {
	entries, idMap := cli.app.findAll()

	cli.printEntries(entries, idMap)
}

func (cli *cli) due() {
	entries, idMap := cli.app.findWhereDueBefore(time.Now())

	cli.printEntries(entries, idMap)
}

func (cli *cli) printEntries(entries []todo, idMap ShortIdMap) {
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
		details := ""
		if len(entry.Details) > 0 {
			details = entry.Details + "\n"
		}
		cli.Resultf("[%s]\n%s\n%s\n%s", blue(entryId), entry.Title, dueFunc(cli.format(entry.Due)), details)
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
	searchFor, timer := ParseTimer(arguments, cli.location)
	if !timer.isEmpty() {
		newDue = timer.Resolve(time.Now())
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
