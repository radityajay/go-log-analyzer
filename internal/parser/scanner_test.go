package parser

import (
	"os"
	"testing"
)

func TestScanFile_Basic(t *testing.T) {
	// Create a temp log file.
	content := `2024-01-15 10:00:00 INFO [gateway] Request started
2024-01-15 10:00:01 ERROR [auth-service] Token expired for user_id=123
2024-01-15 10:00:02 WARN [payment-service] Slow query detected
2024-01-15 10:00:03 ERROR [auth-service] Token expired for user_id=456
2024-01-15 10:00:04 FATAL [gateway] Out of memory
2024-01-15 10:00:05 DEBUG [scheduler] Health check passed
2024-01-15 10:00:06 ERROR [payment-service] Payment gateway timeout
`

	tmpFile, err := os.CreateTemp("", "logalyzer-test-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg := ScanConfig{
		FilePath: tmpFile.Name(),
		Levels:   map[string]bool{"ERROR": true, "FATAL": true},
		Workers:  2,
	}

	result, err := ScanFile(cfg)
	if err != nil {
		t.Fatalf("ScanFile error: %v", err)
	}

	if result.TotalLines != 7 {
		t.Errorf("total lines: got %d, want 7", result.TotalLines)
	}

	if result.MatchedLines != 4 {
		t.Errorf("matched lines: got %d, want 4", result.MatchedLines)
	}

	if result.ByLevel["ERROR"] != 3 {
		t.Errorf("ERROR count: got %d, want 3", result.ByLevel["ERROR"])
	}

	if result.ByLevel["FATAL"] != 1 {
		t.Errorf("FATAL count: got %d, want 1", result.ByLevel["FATAL"])
	}

	if result.BySource["auth-service"] != 2 {
		t.Errorf("auth-service count: got %d, want 2", result.BySource["auth-service"])
	}
}

func TestScanFile_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "logalyzer-empty-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg := ScanConfig{
		FilePath: tmpFile.Name(),
		Levels:   map[string]bool{"ERROR": true},
		Workers:  2,
	}

	result, err := ScanFile(cfg)
	if err != nil {
		t.Fatalf("ScanFile error: %v", err)
	}

	if result.TotalLines != 0 {
		t.Errorf("total lines: got %d, want 0", result.TotalLines)
	}
}

func TestScanFile_SingleWorker(t *testing.T) {
	content := `2024-01-15 10:00:00 ERROR [svc] error one
2024-01-15 10:00:01 ERROR [svc] error two
`
	tmpFile, err := os.CreateTemp("", "logalyzer-single-*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	cfg := ScanConfig{
		FilePath: tmpFile.Name(),
		Levels:   map[string]bool{"ERROR": true},
		Workers:  1,
	}

	result, err := ScanFile(cfg)
	if err != nil {
		t.Fatalf("ScanFile error: %v", err)
	}

	if result.MatchedLines != 2 {
		t.Errorf("matched: got %d, want 2", result.MatchedLines)
	}
}
