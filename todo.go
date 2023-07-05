package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
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

	app := &todoApp{config: createDefaultConfig()}

	var command *string
	var arguments []string
	args := flag.Args()
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
	if *command == "list" {
		app.list()
		os.Exit(0)
	}
	if *command == "show" {
		err := app.show(arguments)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Cannot show. Reason: %s\n", err)
			os.Exit(-2)
		} else {
			os.Exit(0)
		}
	}
	if *command == "add" {
		err := app.add(arguments)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Cannot add. Reason: %s\n", err)
			os.Exit(-3)
		} else {
			os.Exit(0)
		}
	}
}

type todoApp struct {
	config config
}

func (app *todoApp) list() {
	entries := app.scanEntries()

	for _, path := range entries {
		entry := app.readEntryFromFile(path)
		fmt.Printf("%s Title: %s, Details: %s\n", entry.Id, entry.Title, entry.Details)
	}
}

func (app *todoApp) show(arguments []string) error {
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

	entries := app.scanEntries()
	todos := make([]todo, len(entries))
	for i := 0; i < len(entries); i++ {
		todos[i] = app.readEntryFromFile(entries[i])
	}

	var matching *todo

	if findAny && len(todos) > 0 {
		matching = &todos[0]
	}
	if matching == nil {
		for _, entry := range todos {
			if findAny || strings.Contains(strings.ToUpper(entry.Id.String()), strings.ToUpper(searchFor)) {
				matching = &entry
				break
			}
		}
	}
	if matching == nil {
		for _, entry := range todos {
			if strings.Contains(strings.ToUpper(entry.Title), strings.ToUpper(searchFor)) {
				matching = &entry
				break
			}
		}
	}

	if matching == nil {
		fmt.Printf("No path found matching %s\n", searchFor)
		return nil
	}

	fmt.Printf("%s Title: %s, Details: %s\n", matching.Id, matching.Title, matching.Details)
	return nil
}

func (app *todoApp) readEntryFromFile(pathToFile string) todo {
	content, err := os.ReadFile(pathToFile)
	if err != nil {
		log.Fatalf("Failed to read entry from file %s: %s", pathToFile, err)
	}

	var entry todo
	err = yaml.Unmarshal(content, &entry)
	if err != nil {
		log.Fatalf("Failed to parse todo from file %s: %s", pathToFile, err)
	}
	return entry
}

func (app *todoApp) add(arguments []string) error {
	buffer := &bytes.Buffer{}
	for i := 0; i < len(arguments); i++ {
		argument := arguments[i]
		buffer.WriteString(argument)
		if i < len(arguments)-1 {
			buffer.WriteRune(' ')
		}
	}
	todoDir, err := app.findTodoDir()
	if err != nil {
		todoDir = app.createTodoDir()
	}
	title := buffer.String()
	filename := title + ".yml"
	fileContent := todo{Title: title, Id: uuid.New()}
	marshal, err := yaml.Marshal(&fileContent)
	if err != nil {
		log.Fatalf("Failed to write file: %s\n", err)
	}
	err = os.WriteFile(todoDir+"/"+filename, marshal, os.FileMode(0777))
	if err != nil {
		log.Fatalf("Failed to write entry: %s\n", err)
	}
	return nil
}

func (app *todoApp) scanEntries() []string {
	entries := make([]string, 0)
	todoDir, err := app.findTodoDir()
	if err != nil {
		return entries
	}
	files, err := os.ReadDir(todoDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() {
				entries = append(entries, filepath.Join(todoDir, file.Name()))
			}
		}
	}
	return entries
}

func (app *todoApp) findTodoDir() (string, error) {
	stat, err := os.Stat(app.config.todoDir)
	if !os.IsNotExist(err) && stat.IsDir() {
		return app.config.todoDir, nil
	} else if os.IsNotExist(err) {
		return "", errors.New(".todo directory does not exist")
	} else if !stat.IsDir() {
		log.Fatal(".todo is present but not a directory")
	}
	log.Fatal("Error reading .todo directory: ", err)
	return "", nil
}

func (app *todoApp) createTodoDir() string {
	err := os.MkdirAll(app.config.todoDir, os.ModeDir)
	if err != nil {
		log.Fatal("Error writing .todo directory: ", err)
	}
	return app.config.todoDir
}

type config struct {
	todoDir string
}

func createDefaultConfig() config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory: ", err)
	}
	todoDir := homeDir + "/.todo"
	return config{todoDir: todoDir}
}

type todo struct {
	Title   string    `yaml:"title"`
	Details string    `yaml:"details"`
	Id      uuid.UUID `yaml:"id"`
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