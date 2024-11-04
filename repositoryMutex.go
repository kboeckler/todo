package main

import (
	"github.com/google/uuid"
	"sync"
)

type repositoryMutex struct {
	mu        sync.Mutex
	innerRepo repository
}

func newRepositoryMutex(innerRepo repository) *repositoryMutex {
	return &repositoryMutex{mu: sync.Mutex{}, innerRepo: innerRepo}
}

func (r repositoryMutex) readAllEntries() []todo {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.innerRepo.readAllEntries()
}

func (r repositoryMutex) readEntryById(id uuid.UUID) (todo, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.innerRepo.readEntryById(id)
}

func (r repositoryMutex) insertEntry(todo todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.innerRepo.insertEntry(todo)
}

func (r repositoryMutex) updateEntry(todo todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.innerRepo.updateEntry(todo)
}

func (r repositoryMutex) deleteEntry(todo todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.innerRepo.deleteEntry(todo)
}

func (r repositoryMutex) archiveEntry(todo todo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.innerRepo.archiveEntry(todo)
}
