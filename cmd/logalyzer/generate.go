package main

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/radityajayantara/go-log-analyzer/internal/generator"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a dummy log file for testing",
	Long: `Generate creates a realistic dummy log file with configurable size.
Useful for testing and demonstrating logalyzer's scanning capabilities.

Example:
  logalyzer generate --lines=1000000 --output=big.log`,
	RunE: runGenerate,
}

func init() {
	generateCmd.Flags().IntP("lines", "n", 0, "number of log lines to generate (default: 100000)")
	generateCmd.Flags().StringP("output", "o", "", "output file path (default: sample.log)")

	rootCmd.AddCommand(generateCmd)
}

func runGenerate(cmd *cobra.Command, args []string) error {
	lines, _ := cmd.Flags().GetInt("lines")
	output, _ := cmd.Flags().GetString("output")

	// Resolve from config/defaults.
	if lines == 0 {
		lines = viper.GetInt("generate.lines")
		if lines == 0 {
			lines = 100000
		}
	}

	if output == "" {
		output = viper.GetString("generate.output")
		if output == "" {
			output = "sample.log"
		}
	}

	pterm.Println()
	pterm.Info.Printfln("Generating %s lines → %s", pterm.Bold.Sprintf("%d", lines), pterm.Bold.Sprint(output))
	pterm.Println()

	pb, _ := pterm.DefaultProgressbar.
		WithTitle("Generating log file...").
		WithTotal(lines).
		WithShowPercentage(true).
		WithShowElapsedTime(true).
		Start()

	cfg := generator.Config{
		Lines:      lines,
		OutputPath: output,
		OnProgress: func(current, total int) {
			pb.Add(current - pb.Current)
		},
	}

	if err := generator.Generate(cfg); err != nil {
		pb.Stop()
		return fmt.Errorf("generation failed: %w", err)
	}

	pb.Add(lines - pb.Current)
	pb.Stop()

	pterm.Println()
	pterm.Success.Printfln("Generated %d lines → %s", lines, output)
	pterm.Println()

	return nil
}
