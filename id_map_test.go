package main

import (
	"github.com/google/uuid"
	"strings"
	"testing"
)

const (
	ABC = "abc651b1-69ea-4ce3-abac-0f1e829623d5"
	AC  = "ac1651b1-69ea-4ce3-abac-0f1e829623d5"
	BD  = "bdc651b1-69ea-4ce3-abac-0f1e829623d5"
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

func TestCreateIdMap_twoUniqueEntries(t *testing.T) {
	entries := []todo{{Id: uuid.MustParse(ABC)}, {Id: uuid.MustParse(BD)}}
	idMap := CreateIdMap(entries)
	if len(idMap) != 2 {
		t.Errorf("Expected len of idMap to be 2, but was %d", len(idMap))
	}
	aId, abcPresent := idMap[ABC]
	if !abcPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", ABC)
	}
	if abcPresent && !strings.EqualFold("a", aId) {
		t.Errorf("Expected \"%s\" to have \"a\" as a short id, but had \"%s\" instead", ABC, aId)
	}
	bId, bdPresent := idMap[BD]
	if !bdPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", BD)
	}
	if bdPresent && !strings.EqualFold("b", bId) {
		t.Errorf("Expected \"%s\" to have \"b\" as a short id, but had \"%s\" instead", BD, bId)
	}
}

func TestCreateIdMap_twoOverlappingEntries(t *testing.T) {
	entries := []todo{{Id: uuid.MustParse(ABC)}, {Id: uuid.MustParse(AC)}}
	idMap := CreateIdMap(entries)
	if len(idMap) != 2 {
		t.Errorf("Expected len of idMap to be 2, but was %d", len(idMap))
	}
	abId, abcPresent := idMap[ABC]
	if !abcPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", ABC)
	}
	if abcPresent && !strings.EqualFold("ab", abId) {
		t.Errorf("Expected \"%s\" to have \"ab\" as a short id, but had \"%s\" instead", ABC, abId)
	}
	acId, acPresent := idMap[AC]
	if !acPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", AC)
	}
	if acPresent && !strings.EqualFold("ac", acId) {
		t.Errorf("Expected \"%s\" to have \"ac\" as a short id, but had \"%s\" instead", AC, acId)
	}
}

func TestCreateIdMap_threeEntries(t *testing.T) {
	entries := []todo{{Id: uuid.MustParse(ABC)}, {Id: uuid.MustParse(AC)}, {Id: uuid.MustParse(BD)}}
	idMap := CreateIdMap(entries)
	if len(idMap) != 3 {
		t.Errorf("Expected len of idMap to be 3, but was %d", len(idMap))
	}
	abId, abcPresent := idMap[ABC]
	if !abcPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", ABC)
	}
	if abcPresent && !strings.EqualFold("ab", abId) {
		t.Errorf("Expected \"%s\" to have \"ab\" as a short id, but had \"%s\" instead", ABC, abId)
	}
	acId, acPresent := idMap[AC]
	if !acPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", AC)
	}
	if acPresent && !strings.EqualFold("ac", acId) {
		t.Errorf("Expected \"%s\" to have \"ac\" as a short id, but had \"%s\" instead", AC, acId)
	}
	bId, bdPresent := idMap[BD]
	if !bdPresent {
		t.Errorf("Expected \"%s\" to have a short id, but was not present", BD)
	}
	if bdPresent && !strings.EqualFold("b", bId) {
		t.Errorf("Expected \"%s\" to have \"b\" as a short id, but had \"%s\" instead", BD, bId)
	}
}
