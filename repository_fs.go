package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"strings"
)

type repositoryFs struct {
	cfg config
}

func (repo *repositoryFs) readAllEntries() []todo {
	entries := repo.scanEntriesInternal()
	todos := make([]todo, len(entries))
	for i := 0; i < len(entries); i++ {
		todos[i] = repo.readEntryFromFileInternal(entries[i])
	}
	return todos
}

func (repo *repositoryFs) readEntryById(id uuid.UUID) (todo, error) {
	otherIdAsString := id.String()
	entries := repo.scanEntriesInternal()
	for _, entry := range entries {
		todo := repo.readEntryFromFileInternal(entry)
		if strings.EqualFold(todo.Id.String(), otherIdAsString) {
			return todo, nil
		}
	}
	return todo{}, errors.New("no todo present with id " + otherIdAsString)
}

func (repo *repositoryFs) insertEntry(todo todo) error {
	todoDirExists, _ := repo.existsDir(repo.cfg.TodoDir)
	if !todoDirExists {
		repo.createDir(repo.cfg.TodoDir)
	}
	fileName := todo.Title + ".yml"
	filePath := repo.cfg.TodoDir + "/" + fileName
	_, err := os.Stat(filePath)
	if err == nil {
		return errors.New("file already exists")
	}
	todo.filepath = filePath
	repo.writeEntryInternal(todo)
	return nil
}

func (repo *repositoryFs) updateEntry(todo todo) {
	repo.writeEntryInternal(todo)
}

func (repo *repositoryFs) deleteEntry(todo todo) {
	repo.deleteEntryInternal(todo)
}
func (repo *repositoryFs) archiveEntry(todo todo) {
	repo.moveEntryIntoArchiveInternal(todo)
}

func (repo *repositoryFs) writeEntryInternal(todo todo) {
	fileContent, err := yaml.Marshal(&todo)
	if err != nil {
		log.Fatalf("Failed to write file: %s\n", err)
	}
	err = os.WriteFile(todo.filepath, fileContent, os.FileMode(0777))
	if err != nil {
		log.Fatalf("Failed to write entry: %s\n", err)
	}
}

func (repo *repositoryFs) deleteEntryInternal(todo todo) {
	err := os.Remove(todo.filepath)
	if err != nil {
		log.Fatalf("Failed to delete entry: %s\n", err)
	}
}

func (repo *repositoryFs) moveEntryIntoArchiveInternal(todo todo) {
	archiveDir := filepath.Join(repo.cfg.TodoDir, "archive")
	archiveDirExists, _ := repo.existsDir(archiveDir)
	if !archiveDirExists {
		repo.createDir(archiveDir)
	}
	err := os.Rename(todo.filepath, filepath.Join(archiveDir, filepath.Base(todo.filepath)))
	if err != nil {
		log.Fatalf("Failed to move entry: %s\n", err)
	}
}

func (repo *repositoryFs) scanEntriesInternal() []string {
	entries := make([]string, 0)
	todoDirExists, err := repo.existsDir(repo.cfg.TodoDir)
	if !todoDirExists {
		return entries
	}
	files, err := os.ReadDir(repo.cfg.TodoDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && !strings.EqualFold("todo.properties", file.Name()) {
				entries = append(entries, filepath.Join(repo.cfg.TodoDir, file.Name()))
			}
		}
	}
	return entries
}

func (repo *repositoryFs) readEntryFromFileInternal(pathToFile string) todo {
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

func (repo *repositoryFs) existsDir(dirPath string) (bool, error) {
	stat, err := os.Stat(dirPath)
	if !os.IsNotExist(err) && stat.IsDir() {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, fmt.Errorf("%s directory does not exist", dirPath)
	} else if !stat.IsDir() {
		log.Fatalf("%s is present but not a directory", dirPath)
	}
	log.Fatalf("Error reading %s directory: %s", dirPath, err)
	return false, err
}

func (repo *repositoryFs) createDir(dirPath string) {
	err := os.MkdirAll(dirPath, os.FileMode(0777))
	if err != nil {
		log.Fatalf("Error writing %s directory: %s", dirPath, err)
	}
}
