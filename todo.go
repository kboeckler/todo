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
	runAsRestClient := flag.Bool("rest-client", false, "run as rest client - does not do anything when run as server")
	runAsServer := flag.Bool("server", false, "run server instance - additional cli commands will be ignored")
	runInTray := flag.Bool("tray", false, "run in tray - does not do anything when not run as server")
	runAsRestServer := flag.Bool("rest-server", false, "run as rest server - does not do anything when not run as server")
	logFile := flag.String("filename", "", "location of file to append log to - does not do anything when not run as server")
	flag.Usage = usage
	flag.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	config := loadConfig()

	if *runAsServer {
		runServer(logFile, config, runInTray, runAsRestServer)
	} else {
		runCli(config, runAsRestClient)
	}
}

func runServer(logFile *string, config config, runInTray *bool, runAsRestServer *bool) {
	serverFormatter := new(log.JSONFormatter)
	log.SetReportCaller(true)
	log.SetFormatter(serverFormatter)
	if len(*logFile) > 0 {
		log.SetOutput(newFileWriter(*logFile, true))
	}
	log.Debugf("Start server with log level %s", log.GetLevel())

	repo := &repositoryFs{cfg: config}
	app := &appLocal{repo: repo}
	server := server{app: app, cfg: config, runWithTray: *runInTray, runAsRestServer: *runAsRestServer, timeRenderLayout: time.RFC1123}

	server.run()
}

func runCli(config config, runAsRestClient *bool) {
	cliFormatter := new(log.TextFormatter)
	cliFormatter.DisableTimestamp = true
	cliFormatter.DisableLevelTruncation = true
	log.SetFormatter(cliFormatter)

	log.Debugf("Run cli with log level %s", log.GetLevel())

	var app app
	if *runAsRestClient {
		restClient := newRestClient(config.RemoteBaseUrl)
		log.Debugf("Running cli against remote server on BaseUrl '%s'\n", restClient.baseUrl)
		app = newAppRemote(restClient)
	} else {
		repo := &repositoryFs{cfg: config}
		app = &appLocal{repo: repo}
	}
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
	_, _ = fmt.Fprintf(out, "  config\n")
	_, _ = fmt.Fprintf(out, "\tprints the current configuration\n")
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

func showConfig(config config) {
	out := os.Stdout
	_, _ = fmt.Fprintf(out, "Current config:\n")
	_, _ = fmt.Fprintf(out, "  TodoDir=%s\n", config.TodoDir)
	_, _ = fmt.Fprintf(out, "CLI config:\n")
	_, _ = fmt.Fprintf(out, "  EditorCmd=%s\n", config.EditorCmd)
	_, _ = fmt.Fprintf(out, "  RemoteBaseUrl=%s\n", config.RemoteBaseUrl)
	_, _ = fmt.Fprintf(out, "Server config:\n")
	_, _ = fmt.Fprintf(out, "  Tick=%s\n", config.Tick)
	_, _ = fmt.Fprintf(out, "  NotificationCmd=%s\n", config.NotificationCmd)
	_, _ = fmt.Fprintf(out, "  TrayIcon=%s\n", config.TrayIcon)
	_, _ = fmt.Fprintf(out, "  RestBaseHost=%s\n", config.RestBaseHost)
	_, _ = fmt.Fprintf(out, "  RestBasePort=%s\n", config.RestBasePort)
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

type notificationType string

const (
	NotificationTypeNone notificationType = "none"
	NotificationTypeOnce notificationType = "once"
)

type notification struct {
	Type       notificationType `yaml:"type"`
	NotifiedAt time.Time        `yaml:"notifiedAt,omitempty"`
}

type todoModels struct {
	items []todoModel
}

func sorted(items []todoModel) []todoModel {
	sortable := list(items)
	sort.Sort(sortable)
	return sortable.items
}

func list(items []todoModel) *todoModels {
	return &todoModels{items}
}

func (t *todoModels) Len() int {
	return len(t.items)
}

func (t *todoModels) Less(i, j int) bool {
	return !t.items[i].Due.After(t.items[j].Due)
}

func (t *todoModels) Swap(i, j int) {
	tmp := t.items[i]
	t.items[i] = t.items[j]
	t.items[j] = tmp
}
