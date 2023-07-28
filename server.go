package main

import (
	"context"
	"fyne.io/systray"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

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
	if len(server.app.config.NotificationCmd) > 0 {
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
