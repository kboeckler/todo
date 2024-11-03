package main

import (
	"fmt"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

type appRemote struct {
	restClient *restClient
}

func newAppRemote(restClient *restClient) *appRemote {
	return &appRemote{restClient: restClient}
}

func (app appRemote) findAll() ([]todoModel, ShortIdMap) {
	response := TodosResponse{}
	err := app.restClient.doGet("/todos", &response)
	if err != nil {
		log.Errorf("Error requesting all todos: %v\n", err)
	}
	return response.Todos, response.ShortIdMap
}

func (app appRemote) findWhereDueBefore(due time.Time) ([]todoModel, ShortIdMap) {
	response := TodosResponse{}
	searchParams := SearchBody{DueBefore: due}
	err := app.restClient.doPost("/search", searchParams, &response)
	if err != nil {
		log.Errorf("Error finding a todo before '%s': %v\n", due, err)
	}
	return response.Todos, response.ShortIdMap
}

func (app appRemote) findToBeNotifiedByDueBefore(due time.Time) ([]todoModel, ShortIdMap) {
	response := TodosResponse{}
	searchParams := SearchBody{NotifiedBefore: due}
	err := app.restClient.doPost("/search", searchParams, &response)
	if err != nil {
		log.Errorf("Error finding a todo to be notified before '%s': %v\n", due, err)
	}
	return response.Todos, response.ShortIdMap
}

func (app appRemote) find(searchFor string) (*todoModel, string) {
	response := TodosResponse{}
	searchParams := SearchBody{SearchFor: searchFor}
	err := app.restClient.doPost("/search", searchParams, &response)
	if err != nil {
		log.Errorf("Error finding a todo for '%s': %v\n", searchFor, err)
	}
	var responseTodo *todoModel
	responseShortId := ""
	if len(response.Todos) > 0 {
		responseTodo = &response.Todos[0]
		responseShortId = response.ShortIdMap[response.Todos[0].Id.String()]
	}
	return responseTodo, responseShortId
}

func (app appRemote) add(title string, details string, due time.Time) error {
	addParams := AddBody{Title: title, Details: details, Due: due}
	err := app.restClient.doPost("/todos", addParams, nil)
	if err != nil {
		log.Errorf("Error posting a new todo with title '%s': %v\n", title, err)
		return err
	}
	return nil
}

func (app appRemote) delete(todoId uuid.UUID) error {
	err := app.restClient.doDelete(fmt.Sprintf("/todos/%s", todoId))
	if err != nil {
		log.Errorf("Error deleting a todo with the id '%s': %v\n", todoId, err)
		return err
	}
	return nil
}

func (app appRemote) markNotified(todoId uuid.UUID) error {
	err := app.restClient.doPost(fmt.Sprintf("/todos/%s/notified", todoId), nil, nil)
	if err != nil {
		log.Errorf("Error posting a todo as notified with the id '%s': %v\n", todoId, err)
		return err
	}
	return nil
}

func (app appRemote) setNewDue(todoId uuid.UUID, due time.Time) error {
	dueParams := DueBody{Due: due}
	err := app.restClient.doPost(fmt.Sprintf("/todos/%s/due", todoId), dueParams, nil)
	if err != nil {
		log.Errorf("Error posting a new due for a todo with the id '%s': %v\n", todoId, err)
		return err
	}
	return nil
}

func (app appRemote) resolve(todoId uuid.UUID) error {
	err := app.restClient.doPost(fmt.Sprintf("/todos/%s/resolved", todoId), nil, nil)
	if err != nil {
		log.Errorf("Error posting a todo as resolved with the id '%s': %v\n", todoId, err)
		return err
	}
	return nil
}
