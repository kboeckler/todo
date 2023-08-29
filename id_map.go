package main

import (
	"errors"
	"strings"
)

// ShortIdMap maps complete uuids as map keys to shortened strings as map values
type ShortIdMap map[string]string

func CreateIdMap(entries []todo) ShortIdMap {
	idMap := make(ShortIdMap)
	ids := make([]string, len(entries))
	for i := range entries {
		ids[i] = entries[i].Id.String()
	}
	for i, entry := range entries {
		firstUniqueSubstring, _ := findFirstUniqueSubstring(i, ids)
		idMap[entry.Id.String()] = firstUniqueSubstring
	}
	return idMap
}

func findFirstUniqueSubstring(myIndex int, words []string) (string, error) {
	firstUniqueSubstring := ""
	wordLength := len(words[myIndex])
	for i := 0; i < wordLength; i++ {
		firstUniqueSubstring = firstUniqueSubstring + string(words[myIndex][i])
		unique := true
		for j := range words {
			if j != myIndex {
				if strings.Index(words[j], firstUniqueSubstring) == 0 {
					unique = false
					break
				}
			}
		}
		if unique {
			return firstUniqueSubstring, nil
		}
	}
	return "", errors.New("no word is entirely unique")
}
