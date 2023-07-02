package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	debug := flag.Bool("debug", false, "enable debugging messages")

	// Perform command line completion if called from completion library
	complete()

	flag.Parse()

	if *debug {
		fmt.Printf("Running '%s' with parameters:\n", os.Args[0])
		fmt.Printf("  debug:    %v\n", *debug)
	}

	var command *string
	var arguments []string
	args := flag.Args()
	if len(args) > 0 {
		if strings.EqualFold("list", args[0]) {
			command = &args[0]
		}
		if strings.EqualFold("help", args[0]) {
			command = &args[0]
		}
		if strings.EqualFold("show", args[0]) {
			command = &args[0]
			arguments = args[1:]
		}
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
	if *command == "list" {
		list()
		os.Exit(0)
	}
	if *command == "show" {
		err := show(arguments)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Cannot show. Reason: %s\n", err)
		}
	}
}

func list() {
	entries := scanEntries()

	for _, entry := range entries {
		fmt.Println(entry)
	}
}

func scanEntries() []string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory", err)
	}
	entries := make([]string, 0)
	todoDir := homeDir + "/.todo"
	if stat, err := os.Stat(todoDir); !os.IsNotExist(err) {
		if stat.IsDir() {
			files, err := os.ReadDir(todoDir)
			if err == nil {
				for _, file := range files {
					if !file.IsDir() {
						entries = append(entries, file.Name())
					}
				}
			}
		}
	}
	return entries
}

func show(arguments []string) error {
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

	entries := scanEntries()

	for _, entry := range entries {
		if findAny || strings.Contains(strings.ToUpper(entry), strings.ToUpper(searchFor)) {
			fmt.Println(entry)
			return nil
		}
	}

	fmt.Printf("No entry found matching %s\n", searchFor)
	return nil
}

// complete performs bash command line completion for defined flags
// see https://blog.jbowen.dev/2019/11/bash-command-line-completion-with-go/
// Bootstrap completion on terminal by complete -C `pwd`/todo.exe ./todo.exe
func complete() {

	// when Bash calls the command to perform completion it will
	// set several environment variables including COMP_LINE.
	// If this variable is not set, then command is being invoked
	// normally and we can return.
	if _, ok := os.LookupEnv("COMP_LINE"); !ok {
		return
	}

	// Get the current partial word to be completed
	partial := os.Args[2]

	// strip leading '-' from partial, if present
	partial = strings.TrimLeft(partial, "-")

	// Loop through all defined flags and find any that start
	// with partial (or return all flags if partial is empty string)
	// Matching words are returned to Bash via stdout
	flag.VisitAll(func(f *flag.Flag) {
		if partial == "" || strings.HasPrefix(f.Name, partial) {
			fmt.Println("-" + f.Name)
		}
	})

	os.Exit(0)
}
