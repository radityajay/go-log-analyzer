package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "logalyzer",
	Short: "A concurrent log analyzer CLI tool",
	Long: `Logalyzer is a high-performance CLI tool that scans large log files
concurrently, filters by severity level, and presents a rich summary
report in your terminal.

Built with Go concurrency primitives (goroutines, channels, WaitGroup)
to handle 1 GB+ log files efficiently.`,
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ./config/default.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("default")
		viper.SetConfigType("yaml")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("$HOME/.logalyzer")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		// Config file is optional — only warn if explicitly set.
		if cfgFile != "" {
			fmt.Fprintf(os.Stderr, "Warning: cannot read config file: %v\n", err)
		}
	}
}
