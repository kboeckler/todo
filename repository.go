package main

import (
	"errors"
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
	todoDir, err := repo.findTodoDir()
	if err != nil {
		todoDir = repo.createTodoDir()
	}
	filePath := todoDir + "/" + fileName
	_, err = os.Stat(filePath)
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
	todoDir, err := repo.findTodoDir()
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

func (repo *repository) findTodoDir() (string, error) {
	stat, err := os.Stat(repo.config.TodoDir)
	if !os.IsNotExist(err) && stat.IsDir() {
		return repo.config.TodoDir, nil
	} else if os.IsNotExist(err) {
		return "", errors.New(".todo directory does not exist")
	} else if !stat.IsDir() {
		log.Fatal(".todo is present but not a directory")
	}
	log.Fatal("Error reading .todo directory: ", err)
	return "", nil
}

func (repo *repository) createTodoDir() string {
	err := os.MkdirAll(repo.config.TodoDir, os.FileMode(0777))
	if err != nil {
		log.Fatal("Error writing .todo directory: ", err)
	}
	return repo.config.TodoDir
}
