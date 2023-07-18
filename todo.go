package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/google/uuid"
	"github.com/magiconair/properties"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	debug := flag.Bool("debug", false, "enable debugging messages")
	runAsServer := flag.Bool("server", false, "run server instance")

	// Perform command line completion if called from completion library
	complete()

	flag.Parse()

	if *debug {
		fmt.Printf("Running '%s' with parameters:\n", os.Args[0])
		fmt.Printf("  debug:    %v\n", *debug)
	}

	app := &todoApp{}

	if *runAsServer {
		server := server{app}
		server.run()
		os.Exit(0)
	}

	cli := cli{app}
	cli.run()
}

func (cli *cli) run() {
	config := loadConfig()
	cli.app.reloadConfig(config)

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
		cli.list()
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
	if *command == "add" {
		cli.add(arguments)
		os.Exit(0)
	}
	if *command == "due" {
		cli.due()
		os.Exit(0)
	}
}

type cli struct {
	app *todoApp
}

func (cli *cli) list() {
	entries := cli.app.findAll()

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
		fmt.Printf("%s Title: %s, Details: %s, Due: %s\n", entry.Id, entry.Title, entry.Details, entry.Due)
	}
	return nil
}

func (cli *cli) due() {
	entries := cli.app.findWhereDueBefore(time.Now())

	for _, entry := range entries {
		fmt.Printf("%s Title: %s, Details: %s, Due: %s\n", entry.Id, entry.Title, entry.Details, entry.Due)
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
	cli.app.add(title)
}

type todoApp struct {
	config config
}

func (app *todoApp) reloadConfig(config config) {
	app.config = config
}

func (app *todoApp) findAll() []todo {
	return app.readAllEntries()
}

func (app *todoApp) findWhereDueBefore(due time.Time) []todo {
	todos := app.readAllEntries()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) {
			matching = append(matching, entry)
		}
	}

	return matching
}

func (app *todoApp) findWhereDueBeforeAndByNotificationTypeAndNotifiedAtEmpty(due time.Time, notType notificationType) []todo {
	todos := app.readAllEntries()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) && entry.Notification.Type == notType && entry.Notification.NotifiedAt.IsZero() {
			matching = append(matching, entry)
		}
	}

	return matching
}

func (app *todoApp) find(searchFor string) *todo {
	todos := app.readAllEntries()

	var matching *todo

	if matching == nil {
		for _, entry := range todos {
			if strings.Contains(strings.ToUpper(entry.Id.String()), strings.ToUpper(searchFor)) {
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

	return matching
}

func (app *todoApp) markNotified(todoId uuid.UUID) {
	todo, err := app.readEntryById(todoId)
	if err != nil {
		log.Printf("Could not mark todo as notified: %s", err)
	}
	todo.Notification.NotifiedAt = time.Now()
	fmt.Printf("TODO: this todo needs to be persisted: %v\n", todo)
}

func (app *todoApp) add(title string) {
	todoDir, err := app.findTodoDir()
	if err != nil {
		todoDir = app.createTodoDir()
	}
	todo := todo{Title: title, Id: uuid.New(), Due: time.Now().Add(24 * time.Hour), Notification: notification{Type: NotificationTypeOnce}}
	fileContent, err := yaml.Marshal(&todo)
	if err != nil {
		log.Fatalf("Failed to write file: %s\n", err)
	}
	err = os.WriteFile(todoDir+"/"+todo.getRelativeFilepath(), fileContent, os.FileMode(0777))
	if err != nil {
		log.Fatalf("Failed to write entry: %s\n", err)
	}
}

func (app *todoApp) readAllEntries() []todo {
	entries := app.scanEntries()
	todos := make([]todo, len(entries))
	for i := 0; i < len(entries); i++ {
		todos[i] = app.readEntryFromFile(entries[i])
	}
	return todos
}

func (app *todoApp) readEntryById(id uuid.UUID) (todo, error) {
	otherIdAsString := id.String()
	entries := app.scanEntries()
	for _, entry := range entries {
		todo := app.readEntryFromFile(entry)
		if strings.EqualFold(todo.Id.String(), otherIdAsString) {
			return todo, nil
		}
	}
	return todo{}, errors.New("no todo present with if " + otherIdAsString)
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
			if !file.IsDir() && !strings.EqualFold("todo.properties", file.Name()) {
				entries = append(entries, filepath.Join(todoDir, file.Name()))
			}
		}
	}
	return entries
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
	err = entry.validate()
	if err != nil {
		log.Fatalf("Failed to validate todo from file %s: %s", pathToFile, err)
	}
	entry.filepath = pathToFile
	return entry
}

func (app *todoApp) findTodoDir() (string, error) {
	stat, err := os.Stat(app.config.TodoDir)
	if !os.IsNotExist(err) && stat.IsDir() {
		return app.config.TodoDir, nil
	} else if os.IsNotExist(err) {
		return "", errors.New(".todo directory does not exist")
	} else if !stat.IsDir() {
		log.Fatal(".todo is present but not a directory")
	}
	log.Fatal("Error reading .todo directory: ", err)
	return "", nil
}

func (app *todoApp) createTodoDir() string {
	err := os.MkdirAll(app.config.TodoDir, os.FileMode(0777))
	if err != nil {
		log.Fatal("Error writing .todo directory: ", err)
	}
	return app.config.TodoDir
}

type config struct {
	TodoDir         string        `properties:"todoDir,default="`
	Tick            time.Duration `properties:"tick,default=0"`
	NotificationCmd string        `properties:"notification_command,default="`
}

func loadConfig() config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting home directory: ", err)
	}
	todoDir, specified := os.LookupEnv("TODO_USER_HOME")
	if !specified {
		todoDir = homeDir + "/.todo"
	}
	config := config{}
	prop, err := properties.LoadFile(todoDir+"/todo.properties", properties.UTF8)
	if err != nil {
		log.Printf("No config loaded due to error: %s\n, using Defaults", err)
		todoDir = homeDir + "/.todo"
	} else {
		err = prop.Decode(&config)
		if err != nil {
			log.Printf("No config loaded due to error: %s\n, using Defaults", err)
			todoDir = homeDir + "/.todo"
		}
	}
	if len(config.TodoDir) == 0 {
		config.TodoDir = todoDir
	}
	if config.Tick == 0 {
		config.Tick = 1 + time.Second
	}
	if len(config.NotificationCmd) == 0 {
		config.NotificationCmd = "./notification.example"
	}
	return config
}

type todo struct {
	Title        string       `yaml:"title"`
	Details      string       `yaml:"details"`
	Due          time.Time    `yaml:"due,omitempty"`
	Id           uuid.UUID    `yaml:"id"`
	Notification notification `yaml:"notification"`
	filepath     string
}

func (t *todo) validate() error {
	if strings.EqualFold(string(t.Notification.Type), string(NotificationTypeNone)) {
		t.Notification.Type = NotificationTypeNone
	} else if strings.EqualFold(string(t.Notification.Type), string(NotificationTypeOnce)) {
		t.Notification.Type = NotificationTypeOnce
	} else if len(t.Notification.Type) > 0 {
		return errors.New(fmt.Sprintf("notification type %s unknown.", t.Notification.Type))
	}
	return nil
}

func (t *todo) getRelativeFilepath() string {
	if len(t.filepath) > 0 {
		return t.filepath
	} else {
		return t.Title + ".yml"
	}
}

type notificationType string

const (
	NotificationTypeNone notificationType = "none"
	NotificationTypeOnce notificationType = "once"
)

type notification struct {
	Type       notificationType `yaml:"type"`
	NotifiedAt time.Time        `yaml:"notifiedAt,omitempty"`
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

type server struct {
	app *todoApp
}

func (server *server) run() {
	config := loadConfig()
	server.app.reloadConfig(config)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					config = loadConfig()
					server.app.reloadConfig(config)
				case os.Interrupt:
					cancel()
					os.Exit(1)
				}
			case <-ctx.Done():
				log.Printf("Done.")
				os.Exit(1)
			}
		}
	}()

	if err := server.loop(ctx); err != nil {
		log.Fatalf("%s\n", err)
	}

	defer func() {
		cancel()
	}()
}

func (server *server) loop(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.Tick(server.app.config.Tick):
			err := server.handleNotifications()
			if err != nil {
				return err
			}
		}
	}
}

func (server *server) handleNotifications() error {
	todos := server.app.findWhereDueBeforeAndByNotificationTypeAndNotifiedAtEmpty(time.Now(), NotificationTypeOnce)
	for _, todo := range todos {
		cmd := exec.Command(server.app.config.NotificationCmd, "test", todo.Title)
		stdout, err := cmd.Output()
		if err == nil {
			log.Printf("Result of executing notification command: %s", stdout)
			server.app.markNotified(todo.Id)
		}
		if err != nil {
			log.Printf("Error executing notification command: %s\n.", err)
		}
	}
	return nil
}
