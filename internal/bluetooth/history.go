package bluetooth

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// HistoryEvent represents a Bluetooth connect/disconnect event from system logs.
type HistoryEvent struct {
	Timestamp time.Time `json:"timestamp"`
	Device    string    `json:"device"`
	EventType string    `json:"event_type"` // "connected" or "disconnected"
	RawLine   string    `json:"raw_line,omitempty"`
}

// logRunner abstracts log command execution for testing.
var logRunner = func(args ...string) ([]byte, error) {
	return commandRunner("log", args...)
}

// FetchHistory retrieves Bluetooth connect/disconnect events from the system log.
func FetchHistory(duration string) ([]HistoryEvent, error) {
	out, err := logRunner("show",
		"--predicate", `subsystem == "com.apple.bluetooth"`,
		"--style", "compact",
		"--last", duration)
	if err != nil {
		return nil, fmt.Errorf("failed to read system log: %w", err)
	}
	return ParseHistoryEvents(string(out)), nil
}

// connectPattern matches log lines indicating a device connection.
// Examples:
//
//	2024-01-15 10:30:45.123 ... Connected to "AirPods Max"
//	2024-01-15 10:30:45.123 ... connection established for device "AirPods Pro"
var connectPattern = regexp.MustCompile(`(?i)\b(?:connected|connection established|pairing successful)\b`)

// disconnectPattern matches log lines indicating a device disconnection.
var disconnectPattern = regexp.MustCompile(`(?i)\b(?:disconnected|disconnection|link loss)\b`)

// deviceNamePattern extracts a quoted device name from a log line.
var deviceNamePattern = regexp.MustCompile(`"([^"]+)"`)

// timestampPattern matches the compact log timestamp format: YYYY-MM-DD HH:MM:SS.ffffff
var timestampPattern = regexp.MustCompile(`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}\.\d+)`)

// ParseHistoryEvents parses raw system log output into structured Bluetooth events.
func ParseHistoryEvents(logOutput string) []HistoryEvent {
	var events []HistoryEvent
	lines := strings.Split(logOutput, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var eventType string
		if disconnectPattern.MatchString(line) {
			eventType = "disconnected"
		} else if connectPattern.MatchString(line) {
			eventType = "connected"
		} else {
			continue
		}

		ts := parseTimestamp(line)
		device := parseDeviceName(line)

		events = append(events, HistoryEvent{
			Timestamp: ts,
			Device:    device,
			EventType: eventType,
			RawLine:   line,
		})
	}

	return events
}

// parseTimestamp extracts the timestamp from a compact log line.
func parseTimestamp(line string) time.Time {
	matches := timestampPattern.FindStringSubmatch(line)
	if len(matches) < 2 {
		return time.Time{}
	}

	// Try parsing with nanosecond precision first, then microsecond
	for _, layout := range []string{
		"2006-01-02 15:04:05.000000",
		"2006-01-02 15:04:05.000",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.ParseInLocation(layout, matches[1], time.Local); err == nil {
			return t
		}
	}

	return time.Time{}
}

// parseDeviceName extracts a device name from a log line.
// Looks for quoted strings which typically contain the device name.
func parseDeviceName(line string) string {
	matches := deviceNamePattern.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return matches[1]
	}
	return "unknown"
}
