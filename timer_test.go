package main

import (
	"testing"
	"time"
)

func TestParseTimuration_empty(t *testing.T) {
	title, timuration := ParseTimer([]string{"title"}, locationBerlin())
	assertEquals(t, "title", title)
	assertTrue(t, timuration.isEmpty())
}

func TestParseTimuration_onlyDuration(t *testing.T) {
	refTime, _ := time.Parse(time.RFC3339, "2023-11-18T14:00:00+01:00")
	title, timuration := ParseTimer([]string{"title", "2h"}, locationBerlin())
	newTime := timuration.Resolve(refTime)
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T16:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_inDuration(t *testing.T) {
	refTime, _ := time.Parse(time.RFC3339, "2023-11-18T14:00:00+01:00")
	title, timuration := ParseTimer([]string{"title", "in", "2h"}, locationBerlin())
	newTime := timuration.Resolve(refTime)
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T16:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_tomorrow(t *testing.T) {
	refTime, _ := time.Parse(time.RFC3339, "2023-11-18T14:00:00+01:00")
	title, timuration := ParseTimer([]string{"title", "tomorrow"}, locationBerlin())
	newTime := timuration.Resolve(refTime)
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-19T11:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_inNothing(t *testing.T) {
	title, timuration := ParseTimer([]string{"title", "in", "nothing"}, locationBerlin())
	assertEquals(t, "title in nothing", title)
	assertTrue(t, timuration.isEmpty())
}

func TestParseTimuration_onlyTime(t *testing.T) {
	title, timuration := ParseTimer([]string{"title", "2023-11-18 14:00"}, locationBerlin())
	newTime := timuration.Resolve(time.Now())
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T14:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_atTime(t *testing.T) {
	title, timuration := ParseTimer([]string{"title", "at", "2023-11-18 14:00"}, locationBerlin())
	newTime := timuration.Resolve(time.Now())
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T14:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_atTimeInTwoWords(t *testing.T) {
	title, timuration := ParseTimer([]string{"title", "at", "2023-11-18", "14:00"}, locationBerlin())
	newTime := timuration.Resolve(time.Now())
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T14:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_atNothing(t *testing.T) {
	title, timuration := ParseTimer([]string{"title", "at", "nothing"}, locationBerlin())
	assertEquals(t, "title at nothing", title)
	assertTrue(t, timuration.isEmpty())
}
