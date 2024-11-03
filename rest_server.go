package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

type restServer struct {
	app       app
	listeners []restServerListener
}

type restServerListener struct {
	path    string
	handler func(http.ResponseWriter, *http.Request)
}

func listenerOf(path string, handler func(http.ResponseWriter, *http.Request)) restServerListener {
	return restServerListener{path: path, handler: handler}
}

func newRestServer(app app) *restServer {
	rs := &restServer{app: app}
	listeners := make([]restServerListener, 0)
	listeners = append(listeners, listenerOf("/todos", rs.TodosHandler))
	listeners = append(listeners, listenerOf("/todos/{todoId}", rs.TodoHandler))
	listeners = append(listeners, listenerOf("/todos/{todoId}/notified", rs.TodoNotifiedHandler))
	listeners = append(listeners, listenerOf("/todos/{todoId}/resolved", rs.TodoResolvedHandler))
	listeners = append(listeners, listenerOf("/todos/{todoId}/due", rs.TodoDueHandler))
	listeners = append(listeners, listenerOf("/search", rs.SearchHandler))
	rs.listeners = listeners
	return rs
}

type TodosResponse struct {
	Todos      []todoModel `json:"todos"`
	ShortIdMap ShortIdMap  `json:"shortIdMap"`
}

type AddBody struct {
	Title   string    `json:"title"`
	Details string    `json:"details"`
	Due     time.Time `json:"due"`
}

func (rs *restServer) TodosHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Handling incoming request: %s %s\n", r.Method, r.RequestURI)
	method, _, err := rs.resolveMethodAndContentType(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if strings.EqualFold(method, "GET") {
		todos, shortIdMap := rs.app.findAll()
		response := TodosResponse{Todos: todos, ShortIdMap: shortIdMap}
		jsonResponse, err := json.Marshal(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Errorf("Error marshalling JSON: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonResponse)
	} else if strings.EqualFold(method, "POST") {
		addBody := &AddBody{}
		err = rs.parseRequestBody(r.Body, addBody)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		if len(addBody.Title) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("A title must be provided"))
			return
		}
		if addBody.Due.IsZero() {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("A due date must be provided"))
			return
		}
		err := rs.app.add(addBody.Title, addBody.Details, addBody.Due)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func (rs *restServer) TodoHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Handling incoming request: %s %s\n", r.Method, r.RequestURI)
	method, _, err := rs.resolveMethodAndContentType(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	vars := mux.Vars(r)
	todoId, err := uuid.Parse(vars["todoId"])
	if err != nil {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Id is not a valid UUID"))
			return
		}
	}
	if strings.EqualFold(method, "GET") {
		todo, _ := rs.app.find(todoId.String())
		if todo == nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("No todo by the id '%s' found", todoId)))
			log.Debugf("No todo found for requested id '%s'", todoId)
			return
		}
		jsonTodo, err := json.Marshal(todo)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			log.Errorf("Error marshalling JSON: %v", err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(jsonTodo)
	} else if strings.EqualFold(method, "DELETE") {
		err := rs.app.delete(todoId)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func (rs *restServer) TodoNotifiedHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Handling incoming request: %s %s\n", r.Method, r.RequestURI)
	method, _, err := rs.resolveMethodAndContentType(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if !strings.EqualFold(method, "POST") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Method must be 'POST'"))
		return
	}
	vars := mux.Vars(r)
	todoId, err := uuid.Parse(vars["todoId"])
	if err != nil {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Id is not a valid UUID"))
			return
		}
	}
	err = rs.app.markNotified(todoId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (rs *restServer) TodoResolvedHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Handling incoming request: %s %s\n", r.Method, r.RequestURI)
	method, _, err := rs.resolveMethodAndContentType(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if !strings.EqualFold(method, "POST") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Method must be 'POST'"))
		return
	}
	vars := mux.Vars(r)
	todoId, err := uuid.Parse(vars["todoId"])
	if err != nil {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Id is not a valid UUID"))
			return
		}
	}
	err = rs.app.resolve(todoId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type DueBody struct {
	Due time.Time `json:"due"`
}

func (rs *restServer) TodoDueHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Handling incoming request: %s %s\n", r.Method, r.RequestURI)
	method, _, err := rs.resolveMethodAndContentType(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if !strings.EqualFold(method, "POST") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Method must be 'POST'"))
		return
	}
	vars := mux.Vars(r)
	todoId, err := uuid.Parse(vars["todoId"])
	if err != nil {
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Id is not a valid UUID"))
			return
		}
	}
	dueBody := &DueBody{}
	err = rs.parseRequestBody(r.Body, dueBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if dueBody.Due.IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("A due date must be provided"))
		return
	}
	err = rs.app.setNewDue(todoId, dueBody.Due)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type SearchBody struct {
	SearchFor      string    `json:"searchFor"`
	DueBefore      time.Time `json:"dueBefore"`
	NotifiedBefore time.Time `json:"notifiedBefore"`
}

func (rs *restServer) SearchHandler(w http.ResponseWriter, r *http.Request) {
	log.Debugf("Handling incoming request: %s %s\n", r.Method, r.RequestURI)
	method, _, err := rs.resolveMethodAndContentType(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	if !strings.EqualFold(method, "POST") {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Method must be 'POST'"))
		return
	}
	searchBody := &SearchBody{}
	err = rs.parseRequestBody(r.Body, searchBody)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if len(searchBody.SearchFor) == 0 && searchBody.DueBefore.IsZero() && searchBody.NotifiedBefore.IsZero() {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("A search value must be provided"))
		return
	}
	todosResponse := make([]todoModel, 0)
	var shortIdMapResponse ShortIdMap = make(map[string]string)
	if len(searchBody.SearchFor) > 0 {
		todo, shortId := rs.app.find(searchBody.SearchFor)
		if todo != nil {
			todosResponse = append(todosResponse, *todo)
			shortIdMapResponse[(*todo).Id.String()] = shortId
		}
	} else if !searchBody.DueBefore.IsZero() {
		todosResponse, shortIdMapResponse = rs.app.findWhereDueBefore(searchBody.DueBefore)
	} else if !searchBody.NotifiedBefore.IsZero() {
		todosResponse, shortIdMapResponse = rs.app.findToBeNotifiedByDueBefore(searchBody.NotifiedBefore)
	}
	response := TodosResponse{Todos: todosResponse, ShortIdMap: shortIdMapResponse}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		log.Errorf("Error marshalling JSON: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

func (rs *restServer) parseRequestBody(bodyReader io.ReadCloser, parseTarget interface{}) error {
	requestBody, err := io.ReadAll(bodyReader)
	if err != nil {
		log.Errorf("Error reading rest body: %v", err)
		return err
	}
	err = json.Unmarshal(requestBody, parseTarget)
	if err != nil {
		log.Errorf("Error parsing rest body: %v", err)
		return err
	}
	return nil
}

func (rs *restServer) resolveMethodAndContentType(r *http.Request) (string, string, error) {
	method := r.Method
	contentType := r.Header.Get("Content-Type")
	if strings.EqualFold(method, "POST") {
		if len(contentType) == 0 {
			return "", "", errors.New("Content-Type is required")
		}
		if !strings.EqualFold(contentType, "application/json") {
			return "", "", errors.New("Content-Type must be 'application/json'")
		}
	}
	return method, contentType, nil
}
