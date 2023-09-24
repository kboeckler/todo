package main

import (
	"strings"
	"testing"
	"time"
)

func TestFormatRelativeTo_future(t *testing.T) {
	eventTime := "2023-08-23T12:00:00Z"
	relativeTime := "2023-08-19T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "at Wed, 23 Aug 2023", format)
}

func TestFormatRelativeTo_inOneHour(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2023-08-21T11:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "in 1h0m0s", format)
}

func TestFormatRelativeTo_inTwelveHours(t *testing.T) {
	eventTime := "2023-08-21T08:00:00Z"
	relativeTime := "2023-08-20T20:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "in 12h0m0s", format)
}

func TestFormatRelativeTo_inMoreThanTwelveHours_tomorrowAmOneDigit(t *testing.T) {
	eventTime := "2023-08-21T08:00:00Z"
	relativeTime := "2023-08-20T18:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "tomorrow at 08:00", format)
}

func TestFormatRelativeTo_inMoreThanTwelveHours_tomorrowAm(t *testing.T) {
	eventTime := "2023-08-21T10:00:00Z"
	relativeTime := "2023-08-20T20:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "tomorrow at 10:00", format)
}

func TestFormatRelativeTo_inMoreThanTwelveHours_tomorrowPm(t *testing.T) {
	eventTime := "2023-08-21T16:00:00Z"
	relativeTime := "2023-08-20T20:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "tomorrow at 16:00", format)
}

func TestFormatRelativeTo_inMoreThanTwelveHours_tomorrowNextMonth(t *testing.T) {
	eventTime := "2023-09-01T08:00:00Z"
	relativeTime := "2023-08-31T18:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "tomorrow at 08:00", format)
}

func TestFormatRelativeTo_inLessThanTwoDays(t *testing.T) {
	eventTime := "2023-08-22T10:00:00Z"
	relativeTime := "2023-08-20T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "in 2 days", format)
}

func TestFormatRelativeTo_inMoreThanTwoDays(t *testing.T) {
	eventTime := "2023-08-22T10:00:00Z"
	relativeTime := "2023-08-20T08:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "in 2 days", format)
}

func TestFormatRelativeTo_inThreeDays(t *testing.T) {
	eventTime := "2023-08-23T10:00:00Z"
	relativeTime := "2023-08-20T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "in 3 days", format)
}

func TestFormatRelativeTo_past(t *testing.T) {
	eventTime := "2023-08-23T12:00:00Z"
	relativeTime := "2023-08-27T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "since Wed, 23 Aug 2023", format)
}

func TestFormatRelativeTo_forOneHour(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2023-08-21T13:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "for 1h0m0s", format)
}

func TestFormatRelativeTo_forTwelveHours(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2023-08-22T00:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "for 12h0m0s", format)
}

func TestFormatRelativeTo_sinceYesterday(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2023-08-22T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "since yesterday", format)
}

func TestFormatRelativeTo_sinceTwoDays(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2023-08-23T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "for 2 days", format)
}

func TestFormatRelativeTo_sinceThreeDays(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2023-08-24T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, "for 3 days", format)
}

func formatRelativeTo(eventTimeString string, relativeTimeString string) string {
	cli := cli{timeRenderLayout: time.RFC3339, location: locationBerlin()}
	relativeTime, _ := time.Parse(time.RFC3339, relativeTimeString)
	eventTime, _ := time.Parse(time.RFC3339, eventTimeString)
	format := cli.formatRelativeTo(eventTime, relativeTime)
	return format
}

func TestParseTimuration_empty(t *testing.T) {
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title"})
	assertEquals(t, "title", title)
	assertTrue(t, timuration.isEmpty())
}

func TestParseTimuration_onlyDuration(t *testing.T) {
	refTime, _ := time.Parse(time.RFC3339, "2023-11-18T14:00:00+01:00")
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title", "2h"})
	newTime := timuration.CalculateFrom(refTime)
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T16:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_inDuration(t *testing.T) {
	refTime, _ := time.Parse(time.RFC3339, "2023-11-18T14:00:00+01:00")
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title", "in", "2h"})
	newTime := timuration.CalculateFrom(refTime)
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T16:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_inNothing(t *testing.T) {
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title", "in", "nothing"})
	assertEquals(t, "title in nothing", title)
	assertTrue(t, timuration.isEmpty())
}

func TestParseTimuration_onlyTime(t *testing.T) {
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title", "2023-11-18 14:00"})
	newTime := timuration.CalculateFrom(time.Now())
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T14:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_atTime(t *testing.T) {
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title", "at", "2023-11-18 14:00"})
	newTime := timuration.CalculateFrom(time.Now())
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T14:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_atTimeInTwoWords(t *testing.T) {
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title", "at", "2023-11-18", "14:00"})
	newTime := timuration.CalculateFrom(time.Now())
	assertEquals(t, "title", title)
	assertEquals(t, "2023-11-18T14:00:00+01:00", newTime.Format(time.RFC3339))
}

func TestParseTimuration_atNothing(t *testing.T) {
	cli := cli{location: locationBerlin()}
	title, timuration := cli.parseTimuration([]string{"title", "at", "nothing"})
	assertEquals(t, "title at nothing", title)
	assertTrue(t, timuration.isEmpty())
}

func assertEquals(t *testing.T, expected, actual string) {
	if !strings.EqualFold(expected, actual) {
		t.Errorf("Expected format result to be \"%s\", but was \"%s\".", expected, actual)
	}
}

func assertFalse(t *testing.T, actual bool) {
	if false != actual {
		t.Errorf("Expected \"%v\" to be \"%v\", but was not", false, actual)
	}
}

func assertTrue(t *testing.T, actual bool) {
	if true != actual {
		t.Errorf("Expected \"%v\" to be \"%v\", but was not", true, actual)
	}
}

func locationBerlin() *time.Location {
	loc, _ := time.LoadLocation("Europe/Berlin")
	return loc
}
