package parser

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

// ScanResult holds aggregated results from scanning a log file.
type ScanResult struct {
	TotalLines   int64
	MatchedLines int64
	ByLevel      map[string]int64
	BySource     map[string]int64
	TopMessages  []MessageCount
	Entries      []LogEntry
}

// MessageCount pairs a message with its occurrence count.
type MessageCount struct {
	Message string
	Count   int64
}

// chunkResult is the per-worker result collected via channel.
type chunkResult struct {
	TotalLines   int64
	MatchedLines int64
	ByLevel      map[string]int64
	BySource     map[string]int64
	Messages     map[string]int64
	Entries      []LogEntry
}

// ScanConfig holds parameters for the concurrent scan.
type ScanConfig struct {
	FilePath   string
	Levels     map[string]bool
	Workers    int
	OnProgress func(processed int64, total int64)
}

// ScanFile reads the file concurrently and returns aggregated results.
func ScanFile(cfg ScanConfig) (*ScanResult, error) {
	fileInfo, err := os.Stat(cfg.FilePath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat file: %w", err)
	}
	fileSize := fileInfo.Size()

	if fileSize == 0 {
		return &ScanResult{
			ByLevel:  make(map[string]int64),
			BySource: make(map[string]int64),
		}, nil
	}

	workers := cfg.Workers
	if workers < 1 {
		workers = 1
	}

	// Calculate chunk boundaries aligned to newlines.
	boundaries, err := calculateBoundaries(cfg.FilePath, fileSize, workers)
	if err != nil {
		return nil, err
	}

	resultCh := make(chan chunkResult, len(boundaries))
	var wg sync.WaitGroup

	var progressMu sync.Mutex
	var totalProcessed int64

	for _, b := range boundaries {
		wg.Add(1)
		go func(start, end int64) {
			defer wg.Done()
			cr := scanChunk(cfg.FilePath, start, end, cfg.Levels, func(n int64) {
				if cfg.OnProgress != nil {
					progressMu.Lock()
					totalProcessed += n
					current := totalProcessed
					progressMu.Unlock()
					cfg.OnProgress(current, fileSize)
				}
			})
			resultCh <- cr
		}(b[0], b[1])
	}

	wg.Wait()
	close(resultCh)

	return mergeResults(resultCh), nil
}

// boundary is a [start, end) byte range.
type boundary = [2]int64

// calculateBoundaries splits the file into N chunks aligned to line endings.
func calculateBoundaries(path string, size int64, n int) ([]boundary, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	chunkSize := size / int64(n)
	boundaries := make([]boundary, 0, n)
	start := int64(0)

	for i := 0; i < n; i++ {
		end := start + chunkSize
		if i == n-1 {
			end = size
		} else {
			// Align to next newline.
			end, err = alignToNewline(f, end, size)
			if err != nil {
				return nil, err
			}
		}
		if start < end {
			boundaries = append(boundaries, boundary{start, end})
		}
		start = end
	}

	return boundaries, nil
}

// alignToNewline seeks from pos forward to find the next newline.
func alignToNewline(f *os.File, pos, max int64) (int64, error) {
	if pos >= max {
		return max, nil
	}
	if _, err := f.Seek(pos, io.SeekStart); err != nil {
		return 0, err
	}

	buf := make([]byte, 1)
	for cur := pos; cur < max; cur++ {
		if _, err := f.Read(buf); err != nil {
			return max, nil
		}
		if buf[0] == '\n' {
			return cur + 1, nil
		}
	}
	return max, nil
}

// scanChunk reads a byte range of the file and parses matching log lines.
func scanChunk(path string, start, end int64, levels map[string]bool, onBytes func(int64)) chunkResult {
	cr := chunkResult{
		ByLevel:  make(map[string]int64),
		BySource: make(map[string]int64),
		Messages: make(map[string]int64),
	}

	f, err := os.Open(path)
	if err != nil {
		return cr
	}
	defer f.Close()

	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return cr
	}

	reader := io.LimitReader(f, end-start)
	scanner := bufio.NewScanner(reader)

	// Increase buffer for very long lines.
	const maxLineSize = 1024 * 1024 // 1 MB
	scanner.Buffer(make([]byte, 0, 64*1024), maxLineSize)

	var bytesRead int64
	for scanner.Scan() {
		line := scanner.Text()
		bytesRead += int64(len(line)) + 1 // +1 for newline
		cr.TotalLines++

		entry := ParseLine(line)
		if entry == nil {
			continue
		}

		if !entry.MatchesLevels(levels) {
			continue
		}

		cr.MatchedLines++
		cr.ByLevel[entry.Level]++
		cr.BySource[entry.Source]++
		cr.Messages[entry.Message]++
		cr.Entries = append(cr.Entries, *entry)
	}

	if onBytes != nil {
		onBytes(bytesRead)
	}

	return cr
}

// mergeResults aggregates chunk results into a single ScanResult.
func mergeResults(ch <-chan chunkResult) *ScanResult {
	result := &ScanResult{
		ByLevel:  make(map[string]int64),
		BySource: make(map[string]int64),
	}
	allMessages := make(map[string]int64)

	for cr := range ch {
		result.TotalLines += cr.TotalLines
		result.MatchedLines += cr.MatchedLines

		for k, v := range cr.ByLevel {
			result.ByLevel[k] += v
		}
		for k, v := range cr.BySource {
			result.BySource[k] += v
		}
		for k, v := range cr.Messages {
			allMessages[k] += v
		}
		result.Entries = append(result.Entries, cr.Entries...)
	}

	// Build top messages (top 10).
	result.TopMessages = topN(allMessages, 10)

	return result
}

// topN returns the N most frequent items from a frequency map.
func topN(m map[string]int64, n int) []MessageCount {
	items := make([]MessageCount, 0, len(m))
	for k, v := range m {
		items = append(items, MessageCount{Message: k, Count: v})
	}

	// Simple selection sort — fine for small N.
	for i := 0; i < len(items) && i < n; i++ {
		maxIdx := i
		for j := i + 1; j < len(items); j++ {
			if items[j].Count > items[maxIdx].Count {
				maxIdx = j
			}
		}
		items[i], items[maxIdx] = items[maxIdx], items[i]
	}

	if len(items) > n {
		items = items[:n]
	}
	return items
}
