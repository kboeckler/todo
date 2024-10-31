package main

import "github.com/google/uuid"

type repository interface {
	readAllEntries() []todo
	readEntryById(id uuid.UUID) (todo, error)
	insertEntry(todo todo) error
	updateEntry(todo todo)
	deleteEntry(todo todo)
	archiveEntry(todo todo)
}
