package main

import (
	"strings"
	"testing"
	"time"
)

func TestFormatRelativeTo_past(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2099-01-02T15:04:05Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, eventTime, format)
}

func TestFormatRelativeTo_future(t *testing.T) {
	eventTime := "2023-08-21T12:00:00Z"
	relativeTime := "2023-08-18T12:00:00Z"
	format := formatRelativeTo(eventTime, relativeTime)
	assertEquals(t, eventTime, format)
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

func formatRelativeTo(eventTimeString string, relativeTimeString string) string {
	cli := cli{timeRenderLayout: time.RFC3339}
	relativeTime, _ := time.Parse(time.RFC3339, relativeTimeString)
	eventTime, _ := time.Parse(time.RFC3339, eventTimeString)
	format := cli.formatRelativeTo(eventTime, relativeTime)
	return format
}

func assertEquals(t *testing.T, expected, actual string) {
	if !strings.EqualFold(expected, actual) {
		t.Errorf("Expected format result to be \"%s\", but was \"%s\".", expected, actual)
	}
}
