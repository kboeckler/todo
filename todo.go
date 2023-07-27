package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"fyne.io/systray"
	"github.com/google/uuid"
	"github.com/magiconair/properties"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	debug := flag.Bool("debug", false, "enable debugging messages")
	runAsServer := flag.Bool("server", false, "run server instance")
	runInTray := flag.Bool("tray", false, "run in tray - does not do anything when not run as server")

	// Perform command line completion if called from completion library
	complete()

	flag.Parse()

	if *debug {
		fmt.Printf("Running '%s' with parameters:\n", os.Args[0])
		fmt.Printf("  debug:    %v\n", *debug)
	}

	app := &todoApp{repo: &repository{}}

	config := loadConfig()
	app.reloadConfig(config)

	if *runAsServer {
		server := server{app: app}
		if *runInTray {
			server.runSysTray()
		}
		server.run()
		os.Exit(0)
	}

	cli := cli{app}
	cli.run()
}

func (cli *cli) run() {
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
	err := cli.app.add(title)
	if err != nil {
		fmt.Printf("Could not create %s. Maybe this entry already exists?\n", title)
	}
}

type todoApp struct {
	config config
	repo   *repository
}

func (app *todoApp) reloadConfig(config config) {
	app.config = config
	app.repo.config = config
}

func (app *todoApp) findAll() []todo {
	return app.repo.readAllEntries()
}

func (app *todoApp) findWhereDueBefore(due time.Time) []todo {
	todos := app.repo.readAllEntries()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) {
			matching = append(matching, entry)
		}
	}

	return matching
}

func (app *todoApp) findWhereDueBeforeAndByNotificationTypeAndNotifiedAtEmpty(due time.Time, notType notificationType) []todo {
	todos := app.repo.readAllEntries()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) && entry.Notification.Type == notType && entry.Notification.NotifiedAt.IsZero() {
			matching = append(matching, entry)
		}
	}

	return matching
}

func (app *todoApp) find(searchFor string) *todo {
	todos := app.repo.readAllEntries()

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

func (app *todoApp) add(title string) error {
	todo := todo{Title: title, Id: uuid.New(), Due: time.Now().Add(24 * time.Hour), Notification: notification{Type: NotificationTypeOnce}}
	return app.repo.insertEntry(todo, todo.Title+".yml")
}

func (app *todoApp) markNotified(todoId uuid.UUID) {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		log.Printf("Could not mark todo as notified: %s", err)
	}
	todo.Notification.NotifiedAt = time.Now()
	app.repo.updateEntry(todo)
}

type config struct {
	TodoDir         string        `properties:"todoDir,default="`
	Tick            time.Duration `properties:"tick,default=0"`
	NotificationCmd string        `properties:"notification_command,default="`
	TrayIcon        string        `properties:"tray_icon,default="`
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
		log.Printf("No config loaded due to error: %s, using Defaults\n", err)
		todoDir = homeDir + "/.todo"
	} else {
		err = prop.Decode(&config)
		if err != nil {
			log.Printf("No config loaded due to error: %s, using Defaults\n", err)
			todoDir = homeDir + "/.todo"
		}
	}
	if len(config.TodoDir) == 0 {
		config.TodoDir = todoDir
	}
	if config.Tick == 0 {
		config.Tick = 1 * time.Second
	}
	if len(config.NotificationCmd) == 0 {
		config.NotificationCmd = "./notification.example.sh"
	}
	if len(config.TrayIcon) == 0 {
		config.TrayIcon = "todo.png"
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
	app    *todoApp
	ctx    context.Context
	cancel context.CancelFunc
}

func (server *server) run() {
	server.ctx, server.cancel = context.WithCancel(context.Background())

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGHUP)

	defer func() {
		signal.Stop(signalChan)
		server.cancel()
	}()

	go func() {
		for {
			select {
			case s := <-signalChan:
				switch s {
				case syscall.SIGHUP:
					config := loadConfig()
					server.app.reloadConfig(config)
				case os.Interrupt:
					server.cancel()
					os.Exit(1)
				}
			case <-server.ctx.Done():
				log.Printf("Done.\n")
				os.Exit(1)
			}
		}
	}()

	if err := server.loop(server.ctx); err != nil {
		log.Fatalf("%s\n", err)
	}

	defer func() {
		server.cancel()
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
			exitErr, ok := err.(*exec.ExitError)
			debugError := "{}"
			if ok {
				debugError = string(exitErr.Stderr)
			}
			log.Printf("Error executing notification command: %s: %s\n.", err, debugError)
		}
	}
	return nil
}

func (server *server) runSysTray() {
	go systray.Run(server.onReady, server.onExit)
}

func (server *server) onReady() {
	file, err := os.ReadFile(server.app.config.TrayIcon)
	if err != nil {
		log.Printf("Error reading icon from file %s: %s\n", server.app.config.TrayIcon, err)
	}
	systray.SetIcon(file)
	systray.SetTitle("Todo App")
	systray.SetTooltip("Todo App - Server Instance")
	mQuit := systray.AddMenuItem("Quit server instance", "Quit the server instance")

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func (server *server) onExit() {
	server.cancel()
}
