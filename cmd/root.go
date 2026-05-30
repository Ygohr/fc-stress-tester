package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "stress-tester",
	Short: "A CLI stress testing tool for HTTP services",
	Long:  "stress-tester performs concurrent HTTP load tests against a target URL and reports results.",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
