package main

import (
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
	args := flag.Args()
	for _, arg := range args {
		if strings.EqualFold("list", arg) {
			command = &arg
		}
		if strings.EqualFold("help", arg) {
			command = &arg
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

	for _, entry := range entries {
		fmt.Println(entry)
	}
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
