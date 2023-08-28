package main

// maps complete uuids as map keys to shortened strings as map values
type ShortIdMap map[string]string

func CreateIdMap(entries []todo) ShortIdMap {
	idMap := make(ShortIdMap)
	for _, entry := range entries {
		idMap[entry.Id.String()] = entry.Id.String()
	}
	return idMap
}
