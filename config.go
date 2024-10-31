package main

import (
	"fmt"
	"github.com/magiconair/properties"
	"os"
	"time"
)

type config struct {
	TodoDir         string        `properties:"todoDir,default="`
	Tick            time.Duration `properties:"tick,default=0"`
	NotificationCmd string        `properties:"notification_command,default="`
	EditorCmd       string        `properties:"notification_command,default="`
	TrayIcon        string        `properties:"tray_icon,default="`
}

func loadConfig() config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		exitWithError("Error getting home directory: ", err)
	}
	todoDir, specified := os.LookupEnv("TODO_USER_HOME")
	if !specified {
		todoDir = homeDir + "/.todo"
	}
	config := config{}
	prop, err := properties.LoadFile(todoDir+"/todo.properties", properties.UTF8)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "No config loaded due to error: %s, using Defaults\n", err)
		todoDir = homeDir + "/.todo"
	} else {
		err = prop.Decode(&config)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "No config loaded due to error: %s, using Defaults\n", err)
			todoDir = homeDir + "/.todo"
		}
	}
	if len(config.TodoDir) == 0 {
		config.TodoDir = todoDir
	}
	if config.Tick == 0 {
		config.Tick = 1 * time.Second
	}
	if len(config.EditorCmd) == 0 {
		config.EditorCmd = "vim"
	}
	if len(config.TrayIcon) == 0 {
		config.TrayIcon = "todo.png"
	}
	return config
}
