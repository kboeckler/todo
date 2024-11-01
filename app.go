package main

import (
	"github.com/google/uuid"
	"time"
)

type app interface {
	findAll() ([]todoModel, ShortIdMap)
	findWhereDueBefore(due time.Time) ([]todoModel, ShortIdMap)
	findToBeNotifiedByDueBefore(due time.Time) ([]todoModel, ShortIdMap)
	find(searchFor string) (*todoModel, string)
	add(title string, details string, due time.Time) error
	delete(todoId uuid.UUID) error
	markNotified(todoId uuid.UUID) error
	setNewDue(todoId uuid.UUID, due time.Time) error
	resolve(todoId uuid.UUID) error
}

type todoModel struct {
	Title        string            `json:"title"`
	Details      string            `json:"details"`
	Due          time.Time         `json:"due"`
	Id           uuid.UUID         `json:"id"`
	Notification notificationModel `json:"notification"`
	ResolvedAt   time.Time         `json:"resolvedAt"`
}

type notificationModel struct {
	Type       string    `json:"type"`
	NotifiedAt time.Time `json:"notifiedAt"`
}
