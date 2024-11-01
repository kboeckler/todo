package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"os"
	"sort"
	"strings"
	"time"
)

func main() {
	debug := flag.Bool("debug", false, "enable debugging messages")
	runAsServer := flag.Bool("server", false, "run server instance - additional cli commands will be ignored")
	runInTray := flag.Bool("tray", false, "run in tray - does not do anything when not run as server")
	logFile := flag.String("filename", "", "location of file to append log to - does not do anything when not run as server")
	flag.Usage = usage
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
	log.Debugf("Log level is %s", log.GetLevel())

	config := loadConfig()
	repo := &repositoryFs{cfg: config}
	app := &todoApp{repo: repo}

	if *runAsServer {
		runServer(logFile, app, config, runInTray)
	} else {
		runCli(app, config)
	}
}

func runServer(logFile *string, app *todoApp, config config, runInTray *bool) {
	serverFormatter := new(log.JSONFormatter)
	log.SetReportCaller(true)
	log.SetFormatter(serverFormatter)
	if len(*logFile) > 0 {
		log.SetOutput(newFileWriter(*logFile, true))
	}

	server := server{app: app, cfg: config, timeRenderLayout: time.RFC1123}

	if *runInTray {
		server.runSysTray()
	}
	server.run()
}

func runCli(app *todoApp, config config) {
	cliFormatter := new(log.TextFormatter)
	cliFormatter.DisableTimestamp = true
	cliFormatter.DisableLevelTruncation = true
	log.SetFormatter(cliFormatter)

	cli := cli{app, config, output{os.Stdout, os.Stderr}, time.RFC1123, time.Local}

	cli.run(flag.Args())
}

func usage() {
	out := os.Stdout
	_, _ = fmt.Fprintf(out, "Usage: \t%s [-flag] command <argument>\n", os.Args[0])
	_, _ = fmt.Fprintf(out, "\nFlags:\n")
	flag.CommandLine.SetOutput(out)
	flag.PrintDefaults()
	_, _ = fmt.Fprintf(out, "\nCommands:\n")
	_, _ = fmt.Fprintf(out, "  help\n")
	_, _ = fmt.Fprintf(out, "\tprints this help\n")
	_, _ = fmt.Fprintf(out, "  add\n")
	_, _ = fmt.Fprintf(out, "\tadds a new todo\n")
	_, _ = fmt.Fprintf(out, "  list\n")
	_, _ = fmt.Fprintf(out, "\tlists all active todos\n")
	_, _ = fmt.Fprintf(out, "  due\n")
	_, _ = fmt.Fprintf(out, "\tlists all due todos\n")
	_, _ = fmt.Fprintf(out, "  show\n")
	_, _ = fmt.Fprintf(out, "\tprints one todo in detail view\n")
	_, _ = fmt.Fprintf(out, "  del\n")
	_, _ = fmt.Fprintf(out, "\tdeletes an active todo\n")
	_, _ = fmt.Fprintf(out, "  resolve\n")
	_, _ = fmt.Fprintf(out, "\treolves an active todo\n")
	_, _ = fmt.Fprintf(out, "  snooze\n")
	_, _ = fmt.Fprintf(out, "\tsets a new due date for an active todo\n")
}

func exitWithError(v ...any) {
	_, _ = fmt.Fprint(os.Stderr, v...)
	os.Exit(1)
}

type todo struct {
	Title        string       `yaml:"title"`
	Details      string       `yaml:"details"`
	Due          time.Time    `yaml:"due,omitempty"`
	Id           uuid.UUID    `yaml:"id"`
	Notification notification `yaml:"notification"`
	ResolvedAt   time.Time    `yaml:"resolvedAt"`
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

type todos struct {
	items []todo
}

func sorted(items []todo) []todo {
	sortable := list(items)
	sort.Sort(sortable)
	return sortable.items
}

func list(items []todo) *todos {
	return &todos{items}
}

func (t *todos) Len() int {
	return len(t.items)
}

func (t *todos) Less(i, j int) bool {
	return !t.items[i].Due.After(t.items[j].Due)
}

func (t *todos) Swap(i, j int) {
	tmp := t.items[i]
	t.items[i] = t.items[j]
	t.items[j] = tmp
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
