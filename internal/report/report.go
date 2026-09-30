package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/radityajayantara/go-log-analyzer/internal/parser"
)

// Print renders the scan result as colored tables in the terminal.
func Print(result *parser.ScanResult, duration time.Duration, filePath string) {
	pterm.Println()

	// Header
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgCyan)).
		WithTextStyle(pterm.NewStyle(pterm.FgBlack, pterm.Bold)).
		Println("📊 Logalyzer Scan Report")

	pterm.Println()

	// Summary
	summaryData := pterm.TableData{
		{"Metric", "Value"},
		{"File", filePath},
		{"Total Lines Scanned", fmt.Sprintf("%d", result.TotalLines)},
		{"Matched Lines", fmt.Sprintf("%d", result.MatchedLines)},
		{"Scan Duration", duration.Round(time.Millisecond).String()},
		{"Throughput", formatThroughput(result.TotalLines, duration)},
	}

	pterm.DefaultTable.WithHasHeader().WithBoxed().WithData(summaryData).Render()
	pterm.Println()

	// By Level
	if len(result.ByLevel) > 0 {
		pterm.DefaultSection.Println("Matches by Level")
		levelData := pterm.TableData{{"Level", "Count", "Bar"}}

		// Sort levels by severity.
		orderedLevels := []string{"FATAL", "ERROR", "WARN", "INFO", "DEBUG"}
		for _, level := range orderedLevels {
			count, ok := result.ByLevel[level]
			if !ok {
				continue
			}
			bar := makeBar(count, result.MatchedLines, 30)
			styledLevel := styleLevel(level)
			levelData = append(levelData, []string{styledLevel, fmt.Sprintf("%d", count), bar})
		}

		pterm.DefaultTable.WithHasHeader().WithData(levelData).Render()
		pterm.Println()
	}

	// By Source
	if len(result.BySource) > 0 {
		pterm.DefaultSection.Println("Matches by Source")
		sourceData := pterm.TableData{{"Source", "Count"}}

		// Sort by count descending.
		type kv struct {
			Key   string
			Value int64
		}
		var sorted []kv
		for k, v := range result.BySource {
			sorted = append(sorted, kv{k, v})
		}
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Value > sorted[j].Value
		})

		for _, s := range sorted {
			sourceData = append(sourceData, []string{s.Key, fmt.Sprintf("%d", s.Value)})
		}

		pterm.DefaultTable.WithHasHeader().WithData(sourceData).Render()
		pterm.Println()
	}

	// Top Messages
	if len(result.TopMessages) > 0 {
		pterm.DefaultSection.Println("Top Error Messages")
		msgData := pterm.TableData{{"#", "Count", "Message"}}

		for i, m := range result.TopMessages {
			msg := m.Message
			if len(msg) > 80 {
				msg = msg[:77] + "..."
			}
			msgData = append(msgData, []string{
				fmt.Sprintf("%d", i+1),
				fmt.Sprintf("%d", m.Count),
				msg,
			})
		}

		pterm.DefaultTable.WithHasHeader().WithData(msgData).Render()
		pterm.Println()
	}

	// No matches
	if result.MatchedLines == 0 {
		pterm.Warning.Println("No matching log entries found.")
	}
}

// styleLevel applies color to log level strings.
func styleLevel(level string) string {
	switch level {
	case "FATAL":
		return pterm.NewStyle(pterm.FgWhite, pterm.BgRed, pterm.Bold).Sprint(level)
	case "ERROR":
		return pterm.NewStyle(pterm.FgRed, pterm.Bold).Sprint(level)
	case "WARN":
		return pterm.NewStyle(pterm.FgYellow, pterm.Bold).Sprint(level)
	case "INFO":
		return pterm.NewStyle(pterm.FgCyan).Sprint(level)
	case "DEBUG":
		return pterm.NewStyle(pterm.FgGray).Sprint(level)
	default:
		return level
	}
}

// makeBar creates a simple ASCII bar for proportional display.
func makeBar(count, total int64, maxWidth int) string {
	if total == 0 {
		return ""
	}
	ratio := float64(count) / float64(total)
	width := int(ratio * float64(maxWidth))
	if width == 0 && count > 0 {
		width = 1
	}
	pct := ratio * 100
	return fmt.Sprintf("%s %.1f%%", strings.Repeat("█", width), pct)
}

// formatThroughput calculates lines/second.
func formatThroughput(lines int64, d time.Duration) string {
	if d.Seconds() == 0 {
		return "N/A"
	}
	lps := float64(lines) / d.Seconds()
	switch {
	case lps >= 1_000_000:
		return fmt.Sprintf("%.2fM lines/sec", lps/1_000_000)
	case lps >= 1_000:
		return fmt.Sprintf("%.2fK lines/sec", lps/1_000)
	default:
		return fmt.Sprintf("%.0f lines/sec", lps)
	}
}
