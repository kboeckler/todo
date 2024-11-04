package main

import (
	"fmt"
	"github.com/magiconair/properties"
	"os"
	"time"
)

type config struct {
	TodoDir         string        `properties:"todoDir,default="`
	EditorCmd       string        `properties:"editor_command,default="`
	RemoteBaseUrl   string        `properties:"remote_base_url,default="`
	Tick            time.Duration `properties:"tick,default=0"`
	NotificationCmd string        `properties:"notification_command,default="`
	TrayIcon        string        `properties:"tray_icon,default="`
	RestBaseHost    string        `properties:"rest_base_host,default="`
	RestBasePort    string        `properties:"rest_base_port,default="`
}

func loadConfig() config {
	config := readTodoDirAndLoadConfig()
	config = loadCliConfig(config)
	config = loadServerConfig(config)
	return config
}

func readTodoDirAndLoadConfig() config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		exitWithError("Error getting home directory: ", err)
	}
	todoDir, specified := os.LookupEnv("TODO_USER_HOME")
	if !specified {
		todoDir = homeDir + "/.todo"
	}
	resultConfig := config{}
	prop, err := properties.LoadFile(todoDir+"/todo.properties", properties.UTF8)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "No config loaded due to error: %s, using Defaults\n", err)
		todoDir = homeDir + "/.todo"
	} else {
		err = prop.Decode(&resultConfig)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "No config loaded due to error: %s, using Defaults\n", err)
			todoDir = homeDir + "/.todo"
			resultConfig = config{}
		}
	}
	if len(resultConfig.TodoDir) == 0 {
		resultConfig.TodoDir = todoDir
	}
	return resultConfig
}

func loadCliConfig(config config) config {
	if len(config.EditorCmd) == 0 {
		config.EditorCmd = "vim"
	}
	if len(config.RemoteBaseUrl) == 0 {
		config.RemoteBaseUrl = "http://127.0.0.1:8080"
	}
	return config
}

func loadServerConfig(config config) config {
	if config.Tick == 0 {
		config.Tick = 1 * time.Second
	}
	if len(config.NotificationCmd) == 0 {
		config.NotificationCmd = ""
	}
	if len(config.TrayIcon) == 0 {
		config.TrayIcon = "todo.png"
	}
	if len(config.RestBaseHost) == 0 {
		config.RestBaseHost = "0.0.0.0"
	}
	if len(config.RestBasePort) == 0 {
		config.RestBasePort = "8080"
	}
	return config
}
