package main

import (
	"github.com/google/uuid"
	"strings"
	"testing"
)

const (
	ABC = "abc651b1-69ea-4ce3-abac-0f1e829623d5"
)

func TestCreateIdMap_empty(t *testing.T) {
	idMap := CreateIdMap([]todo{})
	if len(idMap) != 0 {
		t.Errorf("Expected len of idMap to be empty, but was %d", len(idMap))
	}
}

func TestCreateIdMap_oneEntry(t *testing.T) {
	entries := []todo{{Id: uuid.MustParse(ABC)}}
	idMap := CreateIdMap(entries)
	if len(idMap) != 1 {
		t.Errorf("Expected len of idMap to be 1, but was %d", len(idMap))
	}
	aId, abcPresent := idMap[ABC]
	if !abcPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", ABC)
	}
	if abcPresent && !strings.EqualFold("a", aId) {
		t.Errorf("Expected \"%s\" to have \"a\" as a short id, but had \"%s\" instead", ABC, aId)
	}
}
