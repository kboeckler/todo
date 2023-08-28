package main

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type todoApp struct {
	config config
	repo   *repository
}

func (app *todoApp) reloadConfig(config config) {
	app.config = config
	app.repo.config = config
}

func (app *todoApp) findAll() ([]todo, ShortIdMap) {
	return app.readAllEntriesAndBuildIdMapInternal()
}

func (app *todoApp) findWhereDueBefore(due time.Time) []todo {
	todos, _ := app.readAllEntriesAndBuildIdMapInternal()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) {
			matching = append(matching, entry)
		}
	}

	return matching
}

func (app *todoApp) findWhereDueBeforeAndByNotificationTypeAndNotifiedAtEmpty(due time.Time, notType notificationType) []todo {
	todos, _ := app.readAllEntriesAndBuildIdMapInternal()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) && entry.Notification.Type == notType && entry.Notification.NotifiedAt.IsZero() {
			matching = append(matching, entry)
		}
	}

	return matching
}

func (app *todoApp) find(searchFor string) *todo {
	todos, _ := app.readAllEntriesAndBuildIdMapInternal()

	var matching *todo

	if matching == nil {
		var firstMatch *todo
		for _, entry := range todos {
			if strings.Index(strings.ToUpper(entry.Id.String()), strings.ToUpper(searchFor)) == 0 {
				if firstMatch != nil {
					firstMatch = nil
					break
				} else {
					firstMatch = &entry
				}
			}
		}
		if firstMatch != nil {
			matching = firstMatch
		}
	}

	if matching == nil {
		for _, entry := range todos {
			if strings.Contains(strings.ToUpper(entry.Id.String()), strings.ToUpper(searchFor)) {
				matching = &entry
				break
			}
		}
	}
	if matching == nil {
		for _, entry := range todos {
			if strings.Contains(strings.ToUpper(entry.Title), strings.ToUpper(searchFor)) {
				matching = &entry
				break
			}
		}
	}

	return matching
}

func (app *todoApp) add(title string, due time.Time) error {
	todo := todo{Title: title, Id: uuid.New(), Due: due, Notification: notification{Type: NotificationTypeOnce}}
	return app.repo.insertEntry(todo, todo.Title+".yml")
}

func (app *todoApp) delete(todoId uuid.UUID) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	app.repo.deleteEntry(todo)
	return nil
}

func (app *todoApp) markNotified(todoId uuid.UUID) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	todo.Notification.NotifiedAt = time.Now()
	app.repo.updateEntry(todo)
	return nil
}

func (app *todoApp) setNewDue(todoId uuid.UUID, due time.Time) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	todo.Due = due
	todo.Notification.NotifiedAt = time.Time{}
	app.repo.updateEntry(todo)
	return nil
}

func (app *todoApp) resolve(todoId uuid.UUID) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	todo.ResolvedAt = time.Now()
	app.repo.updateEntry(todo)
	app.repo.archiveEntry(todo)
	return nil
}

func (app *todoApp) readAllEntriesAndBuildIdMapInternal() ([]todo, ShortIdMap) {
	entries := app.repo.readAllEntries()
	idMap := CreateIdMap(entries)
	return entries, idMap
}
