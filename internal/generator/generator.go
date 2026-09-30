package generator

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"time"
)

var levels = []string{"DEBUG", "INFO", "WARN", "ERROR", "FATAL"}

// levelWeights controls the distribution: mostly INFO/DEBUG, fewer errors.
var levelWeights = []int{30, 40, 15, 12, 3} // must match levels order

var sources = []string{
	"auth-service",
	"payment-service",
	"user-service",
	"notification-service",
	"gateway",
	"scheduler",
	"inventory-service",
}

// messageGenerator is a function that creates a message with random data.
type messageGenerator func(r *rand.Rand) string

var messageTemplates = map[string][]messageGenerator{
	"DEBUG": {
		func(r *rand.Rand) string { return fmt.Sprintf("Processing request with trace_id=tx-%d", r.Intn(99999)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Cache hit for key=user:%d", r.Intn(99999)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Database query completed in %dms", r.Intn(500)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Loading configuration from /etc/app/config-%d.yaml", r.Intn(10)) },
		func(r *rand.Rand) string { return "Health check passed" },
	},
	"INFO": {
		func(r *rand.Rand) string { return fmt.Sprintf("Request completed successfully status=200 duration=%dms", r.Intn(300)) },
		func(r *rand.Rand) string { return fmt.Sprintf("User user_id=%d logged in successfully", r.Intn(99999)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Order order_id=ORD-%d created", r.Intn(99999)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Email sent to recipient=user%d@example.com", r.Intn(9999)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Service started on port=%d", 3000+r.Intn(5000)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Background job job_id=JOB-%d completed", r.Intn(99999)) },
	},
	"WARN": {
		func(r *rand.Rand) string { return fmt.Sprintf("Slow query detected duration=%dms threshold=500ms", 500+r.Intn(2000)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Rate limit approaching for client_id=CL-%d (85%%)", r.Intn(999)) },
		func(r *rand.Rand) string { return fmt.Sprintf("Retry attempt %d/3 for external API call", 1+r.Intn(3)) },
		func(r *rand.Rand) string { return "Memory usage at 78% of allocated limit" },
		func(r *rand.Rand) string { return fmt.Sprintf("Deprecated endpoint called: /api/v1/legacy/%d", r.Intn(20)) },
	},
	"ERROR": {
		func(r *rand.Rand) string { return fmt.Sprintf("Failed to validate token for user_id=%d", r.Intn(99999)) },
		func(r *rand.Rand) string { return "Database connection timeout after 30s" },
		func(r *rand.Rand) string {
			return fmt.Sprintf("Payment gateway returned status=500 for order_id=ORD-%d", r.Intn(99999))
		},
		func(r *rand.Rand) string { return "Failed to send email: SMTP connection refused" },
		func(r *rand.Rand) string { return "Elasticsearch index write failed: bulk indexing error" },
		func(r *rand.Rand) string { return "MongoDB replica set election in progress" },
		func(r *rand.Rand) string { return "Request failed with panic: runtime error: index out of range" },
	},
	"FATAL": {
		func(r *rand.Rand) string { return "Cannot connect to MongoDB: connection refused" },
		func(r *rand.Rand) string { return "TLS certificate expired, shutting down" },
		func(r *rand.Rand) string { return "Out of memory: killing process" },
		func(r *rand.Rand) string { return "Configuration file missing, cannot start service" },
	},
}

// Config holds parameters for log generation.
type Config struct {
	Lines      int
	OutputPath string
	OnProgress func(current, total int)
}

// Generate creates a dummy log file with realistic entries.
func Generate(cfg Config) error {
	f, err := os.Create(cfg.OutputPath)
	if err != nil {
		return fmt.Errorf("cannot create output file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriterSize(f, 256*1024) // 256 KB buffer for fast writes
	defer w.Flush()

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	baseTime := time.Now().Add(-24 * time.Hour) // start from 24h ago

	for i := 0; i < cfg.Lines; i++ {
		// Advance time by 0-2 seconds per line.
		baseTime = baseTime.Add(time.Duration(r.Intn(2000)) * time.Millisecond)

		level := pickLevel(r)
		source := sources[r.Intn(len(sources))]
		message := generateMessage(r, level)

		line := fmt.Sprintf("%s %s [%s] %s\n",
			baseTime.Format("2006-01-02 15:04:05"),
			level,
			source,
			message,
		)

		if _, err := w.WriteString(line); err != nil {
			return fmt.Errorf("write error at line %d: %w", i+1, err)
		}

		if cfg.OnProgress != nil && (i+1)%10000 == 0 {
			cfg.OnProgress(i+1, cfg.Lines)
		}
	}

	// Final progress callback.
	if cfg.OnProgress != nil {
		cfg.OnProgress(cfg.Lines, cfg.Lines)
	}

	return nil
}

// pickLevel selects a log level based on weighted distribution.
func pickLevel(r *rand.Rand) string {
	total := 0
	for _, w := range levelWeights {
		total += w
	}

	n := r.Intn(total)
	cumulative := 0
	for i, w := range levelWeights {
		cumulative += w
		if n < cumulative {
			return levels[i]
		}
	}
	return levels[0]
}

// generateMessage creates a realistic log message for the given level.
func generateMessage(r *rand.Rand, level string) string {
	templates := messageTemplates[level]
	return templates[r.Intn(len(templates))](r)
}
