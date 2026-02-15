package bluetooth

import (
	"fmt"
	"testing"
	"time"
)

func TestParseHistoryEvents(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedCount  int
		expectedTypes  []string
		expectedDevice []string
	}{
		{
			"connected event",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: Connected to "AirPods Max"`,
			1,
			[]string{"connected"},
			[]string{"AirPods Max"},
		},
		{
			"disconnected event",
			`2024-07-15 10:31:00.654321 0x1234 Default com.apple.bluetooth: Disconnected from "AirPods Max"`,
			1,
			[]string{"disconnected"},
			[]string{"AirPods Max"},
		},
		{
			"multiple events",
			`2024-07-15 10:30:00.000000 0x1234 Default com.apple.bluetooth: Connected to "AirPods Max"
2024-07-15 10:30:05.000000 0x1234 Default com.apple.bluetooth: normal log line
2024-07-15 10:31:00.000000 0x1234 Default com.apple.bluetooth: Disconnected from "AirPods Max"
2024-07-15 10:32:00.000000 0x1234 Default com.apple.bluetooth: Connected to "AirPods Pro"`,
			3,
			[]string{"connected", "disconnected", "connected"},
			[]string{"AirPods Max", "AirPods Max", "AirPods Pro"},
		},
		{
			"connection established",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: connection established for "Magic Keyboard"`,
			1,
			[]string{"connected"},
			[]string{"Magic Keyboard"},
		},
		{
			"link loss",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: link loss detected for "Magic Mouse"`,
			1,
			[]string{"disconnected"},
			[]string{"Magic Mouse"},
		},
		{
			"no events",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: scanning started
2024-07-15 10:30:46.123456 0x1234 Default com.apple.bluetooth: service discovered`,
			0,
			nil,
			nil,
		},
		{
			"empty input",
			"",
			0,
			nil,
			nil,
		},
		{
			"no device name",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: device connected`,
			1,
			[]string{"connected"},
			[]string{"unknown"},
		},
		{
			"disconnect takes precedence over connect keyword",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: "AirPods Pro" disconnected from connection`,
			1,
			[]string{"disconnected"},
			[]string{"AirPods Pro"},
		},
		{
			"pairing successful",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: pairing successful with "Magic Trackpad"`,
			1,
			[]string{"connected"},
			[]string{"Magic Trackpad"},
		},
		{
			"case insensitive matching",
			`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: CONNECTED to "AirPods Max"
2024-07-15 10:31:45.123456 0x1234 Default com.apple.bluetooth: DISCONNECTED from "AirPods Max"`,
			2,
			[]string{"connected", "disconnected"},
			[]string{"AirPods Max", "AirPods Max"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := ParseHistoryEvents(tt.input)
			if len(events) != tt.expectedCount {
				t.Fatalf("expected %d events, got %d: %+v", tt.expectedCount, len(events), events)
			}

			for i, ev := range events {
				if i < len(tt.expectedTypes) && ev.EventType != tt.expectedTypes[i] {
					t.Errorf("event[%d]: expected type %q, got %q", i, tt.expectedTypes[i], ev.EventType)
				}
				if i < len(tt.expectedDevice) && ev.Device != tt.expectedDevice[i] {
					t.Errorf("event[%d]: expected device %q, got %q", i, tt.expectedDevice[i], ev.Device)
				}
			}
		})
	}
}

func TestParseTimestamp(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected time.Time
	}{
		{
			"microsecond precision",
			"2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: test",
			time.Date(2024, 7, 15, 10, 30, 45, 123456000, time.Local),
		},
		{
			"millisecond precision",
			"2024-07-15 10:30:45.123 0x1234 Default com.apple.bluetooth: test",
			time.Date(2024, 7, 15, 10, 30, 45, 123000000, time.Local),
		},
		{
			"no timestamp",
			"no timestamp here",
			time.Time{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTimestamp(tt.line)
			if !got.Equal(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestParseDeviceName(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{"quoted name", `Connected to "AirPods Max"`, "AirPods Max"},
		{"no quotes", "Connected to device", "unknown"},
		{"empty quotes", `Connected to ""`, "unknown"},
		{"unicode name", `Connected to "胖胖的大耳机"`, "胖胖的大耳机"},
		{"multiple quotes", `"ignored" then "AirPods"`, "ignored"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDeviceName(tt.line)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestFetchHistory_WithMock(t *testing.T) {
	orig := logRunner
	defer func() { logRunner = orig }()

	logRunner = func(args ...string) ([]byte, error) {
		// Verify correct arguments
		if len(args) < 6 {
			t.Errorf("expected at least 6 args, got %d: %v", len(args), args)
		}
		if args[0] != "show" {
			t.Errorf("expected 'show', got %q", args[0])
		}
		return []byte(`2024-07-15 10:30:45.123456 0x1234 Default com.apple.bluetooth: Connected to "AirPods Max"
2024-07-15 10:31:00.654321 0x1234 Default com.apple.bluetooth: Disconnected from "AirPods Max"`), nil
	}

	events, err := FetchHistory("24h")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].EventType != "connected" {
		t.Errorf("expected connected, got %s", events[0].EventType)
	}
	if events[1].EventType != "disconnected" {
		t.Errorf("expected disconnected, got %s", events[1].EventType)
	}
}

func TestFetchHistory_Error(t *testing.T) {
	orig := logRunner
	defer func() { logRunner = orig }()

	logRunner = func(args ...string) ([]byte, error) {
		return nil, fmt.Errorf("permission denied")
	}

	_, err := FetchHistory("24h")
	if err == nil {
		t.Fatal("expected error")
	}
}
