package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/radityajayantara/go-log-analyzer/internal/parser"
	"github.com/radityajayantara/go-log-analyzer/internal/report"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a log file for errors and warnings",
	Long: `Scan reads a log file concurrently using multiple workers,
filters entries by severity level, and displays a rich summary report.

Example:
  logalyzer scan --file=app.log
  logalyzer scan --file=app.log --level=ERROR,FATAL,WARN --workers=8`,
	RunE: runScan,
}

func init() {
	scanCmd.Flags().StringP("file", "f", "", "path to the log file (required)")
	scanCmd.Flags().StringP("level", "l", "", "log levels to filter (comma-separated, default: ERROR,FATAL)")
	scanCmd.Flags().IntP("workers", "w", 0, "number of concurrent workers (default: 4)")

	scanCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	filePath, _ := cmd.Flags().GetString("file")
	levelStr, _ := cmd.Flags().GetString("level")
	workers, _ := cmd.Flags().GetInt("workers")

	// Resolve from config/defaults.
	if levelStr == "" {
		configLevels := viper.GetStringSlice("scan.levels")
		if len(configLevels) > 0 {
			levelStr = strings.Join(configLevels, ",")
		} else {
			levelStr = "ERROR,FATAL"
		}
	}

	if workers == 0 {
		workers = viper.GetInt("scan.workers")
		if workers == 0 {
			workers = 4
		}
	}

	// Validate file exists.
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", filePath)
	}

	// Build level set.
	levels := make(map[string]bool)
	for _, l := range strings.Split(levelStr, ",") {
		levels[strings.TrimSpace(strings.ToUpper(l))] = true
	}

	// Display scan start.
	pterm.Println()
	pterm.Info.Printfln("Scanning %s", pterm.Bold.Sprint(filePath))
	pterm.Info.Printfln("Levels: %s | Workers: %d", pterm.Bold.Sprint(levelStr), workers)
	pterm.Println()

	// Progress bar.
	pb, _ := pterm.DefaultProgressbar.
		WithTitle("Scanning log file...").
		WithTotal(100).
		WithShowPercentage(true).
		WithShowElapsedTime(true).
		Start()

	startTime := time.Now()

	cfg := parser.ScanConfig{
		FilePath: filePath,
		Levels:   levels,
		Workers:  workers,
		OnProgress: func(processed, total int64) {
			if total > 0 {
				pct := int(float64(processed) / float64(total) * 100)
				if pct > 100 {
					pct = 100
				}
				pb.Add(pct - pb.Current)
			}
		},
	}

	result, err := parser.ScanFile(cfg)
	if err != nil {
		pb.Stop()
		return fmt.Errorf("scan failed: %w", err)
	}

	// Complete progress bar.
	pb.Add(100 - pb.Current)
	pb.Stop()

	duration := time.Since(startTime)

	// Render report.
	report.Print(result, duration, filePath)

	return nil
}
