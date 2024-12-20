package main

import (
	"context"
	"fmt"
	"fyne.io/systray"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type server struct {
	app              app
	cfg              config
	runWithTray      bool
	runAsRestServer  bool
	ctx              context.Context
	cancel           context.CancelFunc
	timeRenderLayout string
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
					newConfig := loadConfig()
					newRepo := newRepositoryMutex(newRepositoryFs(newConfig))
					newApp := &appLocal{repo: newRepo}
					server.cfg = newConfig
					server.app = newApp
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

	if server.runWithTray {
		go server.runSysTray()
	}

	if server.runAsRestServer {
		go server.runRestServer()
	}

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
		case <-time.Tick(server.cfg.Tick):
			err := server.handleNotifications()
			if err != nil {
				return err
			}
		}
	}
}

func (server *server) handleNotifications() error {
	if len(server.cfg.NotificationCmd) > 0 {
		todos, _ := server.app.findToBeNotifiedByDueBefore(time.Now())
		for _, todo := range todos {
			cmd := exec.Command(server.cfg.NotificationCmd, todo.Title, server.renderNotificationText(todo))
			log.Debugf("Calling notification command: %s", cmd)
			stdout, err := cmd.Output()
			if err == nil {
				log.Debugf("Result of executing notification command: %s", stdout)
				err := server.app.markNotified(todo.Id)
				if err != nil {
					log.Errorf("Could not mark as notified: %s %s: %s", todo.Id, todo.Title, err)
				}
			}
			if err != nil {
				exitErr, ok := err.(*exec.ExitError)
				debugError := "{}"
				if ok {
					debugError = string(exitErr.Stderr)
				}
				log.Errorf("Error executing notification command: %s: Stdout: %s. DebugErr: %s.", err, stdout, debugError)
			}
		}
	}
	return nil
}

func (server *server) renderNotificationText(todo todoModel) string {
	return fmt.Sprintf("%s\n%s\n%s", todo.Title, todo.Due.Format(server.timeRenderLayout), todo.Details)
}

func (server *server) runSysTray() {
	log.Debugf("Running in tray now")
	systray.Run(server.onReady, server.onExit)
}

func (server *server) onReady() {
	file, err := os.ReadFile(server.cfg.TrayIcon)
	if err != nil {
		log.Errorf("Error reading icon from file '%s': %v\n", server.cfg.TrayIcon, err)
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

func (server *server) runRestServer() {
	restServer := newRestServer(server.app)
	r := mux.NewRouter()
	srv := &http.Server{
		Addr: fmt.Sprintf("%s:%s", server.cfg.RestBaseHost, server.cfg.RestBasePort),
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}
	for _, listener := range restServer.listeners {
		r.HandleFunc(listener.path, listener.handler)
	}
	log.Debugf("Running rest server on Address '%s'\n", srv.Addr)
	srv.ListenAndServe()
}
