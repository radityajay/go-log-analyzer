# 🔍 Logalyzer

A high-performance, concurrent log analyzer CLI tool built with Go.

Logalyzer scans large log files (1 GB+) using concurrent workers, filters entries by severity level, and presents a rich, color-coded summary report directly in your terminal.

## Features

- **Concurrent Scanning** — Splits files into byte-offset chunks and processes them in parallel using goroutines
- **Configurable Filtering** — Filter by log level (DEBUG, INFO, WARN, ERROR, FATAL)
- **Rich Terminal Output** — Color-coded tables, progress bars, and throughput metrics via [pterm](https://github.com/pterm/pterm)
- **Log Generator** — Built-in command to generate realistic dummy log files for testing
- **YAML Configuration** — Powered by [Viper](https://github.com/spf13/viper) for flexible config management

## Quick Start

```bash
# Build
go build -o logalyzer ./cmd/logalyzer/

# Generate a test log file (100K lines)
./logalyzer generate --lines=100000 --output=sample.log

# Scan for errors
./logalyzer scan --file=sample.log

# Scan with custom levels and worker count
./logalyzer scan --file=sample.log --level=ERROR,FATAL,WARN --workers=8
```

## Installation

```bash
git clone https://github.com/radityajayantara/go-log-analyzer.git
cd go-log-analyzer
go build -o logalyzer ./cmd/logalyzer/
```

## Commands

### `scan` — Analyze a log file

```bash
logalyzer scan --file=<path> [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--file` | `-f` | (required) | Path to the log file |
| `--level` | `-l` | `ERROR,FATAL` | Log levels to filter (comma-separated) |
| `--workers` | `-w` | `4` | Number of concurrent workers |
| `--config` | | `./config/default.yaml` | Path to config file |

### `generate` — Create a dummy log file

```bash
logalyzer generate [flags]
```

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--lines` | `-n` | `100000` | Number of log lines to generate |
| `--output` | `-o` | `sample.log` | Output file path |

## Supported Log Format

```
2024-01-15 10:23:45 ERROR [auth-service] Failed to validate token for user_id=882
```

Pattern: `<YYYY-MM-DD HH:MM:SS> <LEVEL> [<source>] <message>`

## How It Works

1. **File Chunking** — The file is divided into N byte-offset chunks (where N = number of workers), with boundaries aligned to newline characters to avoid splitting lines
2. **Parallel Scanning** — Each chunk is assigned to a goroutine that independently parses and filters log entries
3. **Result Aggregation** — Results from all workers are collected via channels and merged into a single report
4. **Report Rendering** — The aggregated data is rendered as colored tables showing:
   - Summary statistics (total lines, matches, duration, throughput)
   - Breakdown by log level with proportional bars
   - Breakdown by source service
   - Top 10 most frequent error messages

## Architecture

```
logalyzer scan --file=app.log --workers=4
         │
         ▼
┌─────────────────┐
│  Cobra Command   │
│  (scan.go)       │
└────────┬────────┘
         │
         ▼
┌─────────────────┐     ┌──────────┐  ┌──────────┐
│  ScanFile()      │────▶│ Worker 1 │  │ Worker 2 │
│  (scanner.go)    │     │ [0, 250MB)│  │[250, 500MB)│
│                  │     └─────┬────┘  └─────┬────┘
│  Byte-offset     │     ┌─────┴────┐  ┌─────┴────┐
│  chunking        │     │ Worker 3 │  │ Worker 4 │
│                  │     │[500,750MB)│  │[750MB,1GB)│
└────────┬────────┘     └─────┬────┘  └─────┬────┘
         │                    │              │
         ▼                    ▼              ▼
┌─────────────────┐     ┌────────────────────┐
│  mergeResults()  │◀────│   channel results  │
└────────┬────────┘     └────────────────────┘
         │
         ▼
┌─────────────────┐
│  report.Print()  │
│  (pterm tables)  │
└─────────────────┘
```

## Project Structure

```
go-log-analyzer/
├── cmd/logalyzer/
│   ├── main.go          # Entry point
│   ├── root.go          # Root command + Viper config
│   ├── scan.go          # Scan command
│   └── generate.go      # Generate command
├── internal/
│   ├── parser/
│   │   ├── parser.go    # Log line parsing
│   │   ├── scanner.go   # Concurrent chunk scanning
│   │   ├── parser_test.go
│   │   └── scanner_test.go
│   ├── generator/
│   │   └── generator.go # Dummy log generator
│   └── report/
│       └── report.go    # Terminal report rendering
├── config/
│   └── default.yaml     # Default configuration
├── go.mod
└── go.sum
```

## Configuration

Create a `config/default.yaml` or `~/.logalyzer/default.yaml`:

```yaml
scan:
  levels:
    - ERROR
    - FATAL
  workers: 4

generate:
  lines: 100000
  output: sample.log
```

## Tech Stack

| Component | Library |
|-----------|---------|
| CLI Framework | [spf13/cobra](https://github.com/spf13/cobra) |
| Configuration | [spf13/viper](https://github.com/spf13/viper) |
| Terminal UI | [pterm/pterm](https://github.com/pterm/pterm) |
| Concurrency | Go stdlib (goroutines, sync.WaitGroup, channels) |

## License

MIT
