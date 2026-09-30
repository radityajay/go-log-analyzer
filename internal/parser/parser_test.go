package parser

import (
	"testing"
	"time"
)

func TestParseLine_ValidEntry(t *testing.T) {
	line := "2024-01-15 10:23:45 ERROR [auth-service] Failed to validate token for user_id=882"
	entry := ParseLine(line)

	if entry == nil {
		t.Fatal("expected non-nil entry")
	}

	expected := time.Date(2024, 1, 15, 10, 23, 45, 0, time.UTC)
	if !entry.Timestamp.Equal(expected) {
		t.Errorf("timestamp: got %v, want %v", entry.Timestamp, expected)
	}
	if entry.Level != "ERROR" {
		t.Errorf("level: got %q, want %q", entry.Level, "ERROR")
	}
	if entry.Source != "auth-service" {
		t.Errorf("source: got %q, want %q", entry.Source, "auth-service")
	}
	if entry.Message != "Failed to validate token for user_id=882" {
		t.Errorf("message: got %q", entry.Message)
	}
}

func TestParseLine_AllLevels(t *testing.T) {
	levels := []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}
	for _, level := range levels {
		line := "2024-01-15 10:23:45 " + level + " [svc] msg"
		entry := ParseLine(line)
		if entry == nil {
			t.Errorf("level %s: expected non-nil entry", level)
			continue
		}
		if entry.Level != level {
			t.Errorf("level: got %q, want %q", entry.Level, level)
		}
	}
}

func TestParseLine_InvalidLines(t *testing.T) {
	cases := []struct {
		name string
		line string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"no timestamp", "ERROR [auth] something"},
		{"bad level", "2024-01-15 10:23:45 TRACE [auth] msg"},
		{"no brackets", "2024-01-15 10:23:45 ERROR auth msg"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if entry := ParseLine(tc.line); entry != nil {
				t.Errorf("expected nil for %q, got %+v", tc.line, entry)
			}
		})
	}
}

func TestMatchesLevels(t *testing.T) {
	entry := &LogEntry{Level: "ERROR"}
	levels := map[string]bool{"ERROR": true, "FATAL": true}

	if !entry.MatchesLevels(levels) {
		t.Error("ERROR should match")
	}

	levels2 := map[string]bool{"INFO": true}
	if entry.MatchesLevels(levels2) {
		t.Error("ERROR should not match INFO-only set")
	}
}
