package main

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type repository struct {
	config config
}

func (repo *repository) insertEntry(todo todo, fileName string) error {
	todoDirExists, _ := repo.existsDir(repo.config.TodoDir)
	if !todoDirExists {
		repo.createDir(repo.config.TodoDir)
	}
	filePath := repo.config.TodoDir + "/" + fileName
	_, err := os.Stat(filePath)
	if err == nil {
		return errors.New("file already exists")
	}
	todo.filepath = filePath
	repo.writeEntry(todo)
	return nil
}

func (repo *repository) updateEntry(todo todo) {
	repo.writeEntry(todo)
}

func (repo *repository) deleteEntry(todo todo) {
	repo.deleteEntryInternal(todo)
}
func (repo *repository) archiveEntry(todo todo) {
	repo.moveEntryIntoArchive(todo)
}

func (repo *repository) writeEntry(todo todo) {
	fileContent, err := yaml.Marshal(&todo)
	if err != nil {
		log.Fatalf("Failed to write file: %s\n", err)
	}
	err = os.WriteFile(todo.filepath, fileContent, os.FileMode(0777))
	if err != nil {
		log.Fatalf("Failed to write entry: %s\n", err)
	}
}

func (repo *repository) deleteEntryInternal(todo todo) {
	err := os.Remove(todo.filepath)
	if err != nil {
		log.Fatalf("Failed to delete entry: %s\n", err)
	}
}

func (repo *repository) moveEntryIntoArchive(todo todo) {
	archiveDir := filepath.Join(repo.config.TodoDir, "archive")
	archiveDirExists, _ := repo.existsDir(archiveDir)
	if !archiveDirExists {
		repo.createDir(archiveDir)
	}
	err := os.Rename(todo.filepath, filepath.Join(archiveDir, filepath.Base(todo.filepath)))
	if err != nil {
		log.Fatalf("Failed to move entry: %s\n", err)
	}
}

func (repo *repository) readAllEntries() []todo {
	entries := repo.scanEntries()
	todos := make([]todo, len(entries))
	for i := 0; i < len(entries); i++ {
		todos[i] = repo.readEntryFromFile(entries[i])
	}
	return todos
}

func (repo *repository) readEntryById(id uuid.UUID) (todo, error) {
	otherIdAsString := id.String()
	entries := repo.scanEntries()
	for _, entry := range entries {
		todo := repo.readEntryFromFile(entry)
		if strings.EqualFold(todo.Id.String(), otherIdAsString) {
			return todo, nil
		}
	}
	return todo{}, errors.New("no todo present with id " + otherIdAsString)
}

func (repo *repository) scanEntries() []string {
	entries := make([]string, 0)
	todoDirExists, err := repo.existsDir(repo.config.TodoDir)
	if !todoDirExists {
		return entries
	}
	files, err := os.ReadDir(repo.config.TodoDir)
	if err == nil {
		for _, file := range files {
			if !file.IsDir() && !strings.EqualFold("todo.properties", file.Name()) {
				entries = append(entries, filepath.Join(repo.config.TodoDir, file.Name()))
			}
		}
	}
	return entries
}

func (repo *repository) readEntryFromFile(pathToFile string) todo {
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

func (repo *repository) existsDir(dirPath string) (bool, error) {
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

func (repo *repository) createDir(dirPath string) {
	err := os.MkdirAll(dirPath, os.FileMode(0777))
	if err != nil {
		log.Fatalf("Error writing %s directory: %s", dirPath, err)
	}
}
