package parser

import (
	"regexp"
	"strings"
	"time"
)

// LogEntry represents a single parsed log line.
type LogEntry struct {
	Timestamp time.Time
	Level     string
	Source    string
	Message   string
	Raw       string
}

// logPattern matches: 2024-01-15 10:23:45 ERROR [auth-service] message...
var logPattern = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2})\s+(DEBUG|INFO|WARN|ERROR|FATAL)\s+\[([^\]]+)\]\s+(.+)$`,
)

const timeLayout = "2006-01-02 15:04:05"

// ParseLine parses a single log line into a LogEntry.
// Returns nil if the line does not match the expected format.
func ParseLine(line string) *LogEntry {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil
	}

	matches := logPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	ts, err := time.Parse(timeLayout, matches[1])
	if err != nil {
		return nil
	}

	return &LogEntry{
		Timestamp: ts,
		Level:     matches[2],
		Source:    matches[3],
		Message:   matches[4],
		Raw:       line,
	}
}

// MatchesLevels checks whether the entry's level is in the given set.
func (e *LogEntry) MatchesLevels(levels map[string]bool) bool {
	return levels[e.Level]
}
