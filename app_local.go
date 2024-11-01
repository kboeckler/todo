package main

import (
	"github.com/google/uuid"
	"strings"
	"time"
)

type appLocal struct {
	repo repository
}

func (app *appLocal) findAll() ([]todoModel, ShortIdMap) {
	return mapTodosWithIdMap(app.readAllEntriesAndBuildIdMapInternal())
}

func (app *appLocal) findWhereDueBefore(due time.Time) ([]todoModel, ShortIdMap) {
	todos, idMap := app.readAllEntriesAndBuildIdMapInternal()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) {
			matching = append(matching, entry)
		}
	}

	return mapTodosWithIdMap(matching, idMap)
}

func (app *appLocal) findToBeNotifiedByDueBefore(due time.Time) ([]todoModel, ShortIdMap) {
	todos, idMap := app.readAllEntriesAndBuildIdMapInternal()

	matching := make([]todo, 0)

	for _, entry := range todos {
		if entry.Due.Before(due) && entry.Notification.Type == NotificationTypeOnce && entry.Notification.NotifiedAt.IsZero() {
			matching = append(matching, entry)
		}
	}

	return mapTodosWithIdMap(matching, idMap)
}

func (app *appLocal) find(searchFor string) (*todoModel, string) {
	todos, idMap := app.readAllEntriesAndBuildIdMapInternal()

	var matching *todo

	if matching == nil {
		var todoId string
		for original, short := range idMap {
			if strings.EqualFold(strings.ToUpper(short), strings.ToUpper(searchFor)) {
				todoId = original
				break
			}
		}
		if len(todoId) > 0 {
			for _, entry := range todos {
				if strings.EqualFold(entry.Id.String(), todoId) {
					matching = &entry
					break
				}
			}
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

	shortId := ""
	if matching != nil {
		shortId = idMap[matching.Id.String()]
	}

	return mapTodoWithShortId(matching, shortId)
}

func (app *appLocal) add(title string, details string, due time.Time) error {
	todo := todo{Title: title, Details: details, Id: uuid.New(), Due: due, Notification: notification{Type: NotificationTypeOnce}}
	return app.repo.insertEntry(todo)
}

func (app *appLocal) delete(todoId uuid.UUID) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	app.repo.deleteEntry(todo)
	return nil
}

func (app *appLocal) markNotified(todoId uuid.UUID) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	todo.Notification.NotifiedAt = time.Now()
	app.repo.updateEntry(todo)
	return nil
}

func (app *appLocal) setNewDue(todoId uuid.UUID, due time.Time) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	todo.Due = due
	todo.Notification.NotifiedAt = time.Time{}
	app.repo.updateEntry(todo)
	return nil
}

func (app *appLocal) resolve(todoId uuid.UUID) error {
	todo, err := app.repo.readEntryById(todoId)
	if err != nil {
		return err
	}
	todo.ResolvedAt = time.Now()
	app.repo.updateEntry(todo)
	app.repo.archiveEntry(todo)
	return nil
}

func (app *appLocal) readAllEntriesAndBuildIdMapInternal() ([]todo, ShortIdMap) {
	entries := app.repo.readAllEntries()
	idMap := CreateIdMap(entries)
	return entries, idMap
}

func mapTodosWithIdMap(todos []todo, shortIdMap ShortIdMap) ([]todoModel, ShortIdMap) {
	return mapTodos(todos), shortIdMap
}

func mapTodoWithShortId(todo *todo, shortId string) (*todoModel, string) {
	if todo == nil {
		return nil, ""
	}
	todoModel := mapTodo(*todo)
	return &todoModel, shortId
}

func mapTodos(todos []todo) []todoModel {
	res := make([]todoModel, 0, len(todos))
	for _, t := range todos {
		res = append(res, mapTodo(t))
	}
	return res
}

func mapTodo(todo todo) todoModel {
	return todoModel{
		Title:        todo.Title,
		Details:      todo.Details,
		Due:          todo.Due,
		Id:           todo.Id,
		Notification: mapNotification(todo.Notification),
		ResolvedAt:   todo.ResolvedAt,
	}
}

func mapNotification(notification notification) notificationModel {
	return notificationModel{
		Type:       mapNotificationType(notification.Type),
		NotifiedAt: notification.NotifiedAt,
	}
}

func mapNotificationType(notificationType notificationType) string {
	switch notificationType {
	case NotificationTypeNone:
		return "none"
	case NotificationTypeOnce:
		return "once"
	}
	return ""
}
