package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"

	"github.com/Ygohr/fc-stress-tester/infrastructure/http"
	"github.com/Ygohr/fc-stress-tester/internal/service/loadtest"
	"github.com/spf13/cobra"
)

var (
	flagURL         string
	flagRequests    int
	flagConcurrency int
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a load test against a target URL",
	RunE:  runLoadTest,
}

func init() {
	runCmd.Flags().StringVar(&flagURL, "url", "", "Target URL to load test (required)")
	runCmd.Flags().IntVar(&flagRequests, "requests", 0, "Total number of requests to perform (required)")
	runCmd.Flags().IntVar(&flagConcurrency, "concurrency", 0, "Number of concurrent workers (required)")

	_ = runCmd.MarkFlagRequired("url")
	_ = runCmd.MarkFlagRequired("requests")
	_ = runCmd.MarkFlagRequired("concurrency")

	rootCmd.AddCommand(runCmd)

	rootCmd.Flags().StringVar(&flagURL, "url", "", "Target URL to load test")
	rootCmd.Flags().IntVar(&flagRequests, "requests", 0, "Total number of requests to perform")
	rootCmd.Flags().IntVar(&flagConcurrency, "concurrency", 0, "Number of concurrent workers")
	rootCmd.RunE = runLoadTest
}

func runLoadTest(cmd *cobra.Command, _ []string) error {
	if err := validateFlags(); err != nil {
		return err
	}

	client := http.NewHTTPClient()
	svc := loadtest.NewService(client)

	cfg := loadtest.Config{
		URL:         flagURL,
		Requests:    flagRequests,
		Concurrency: flagConcurrency,
	}

	report, err := svc.Execute(context.Background(), cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load test failed: %v\n", err)
		os.Exit(1)
	}

	printReport(report)
	return nil
}

func validateFlags() error {
	if flagURL == "" {
		return fmt.Errorf("--url is required and must not be empty")
	}

	parsed, err := url.ParseRequestURI(flagURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return fmt.Errorf("--url must be a valid HTTP or HTTPS URL, got: %q", flagURL)
	}

	if flagRequests <= 0 {
		return fmt.Errorf("--requests must be greater than 0, got: %d", flagRequests)
	}

	if flagConcurrency <= 0 {
		return fmt.Errorf("--concurrency must be greater than 0, got: %d", flagConcurrency)
	}

	return nil
}

func printReport(r loadtest.Report) {
	fmt.Println()
	fmt.Println("========== Load Test Report ==========")
	fmt.Printf("Total execution time:  %s\n", r.TotalDuration.Round(1000000))
	fmt.Printf("Total requests:        %d\n", r.TotalRequests)
	fmt.Printf("HTTP 200:              %d\n", r.SuccessCount)
	fmt.Printf("Request errors:        %d\n", r.ErrorCount)

	if len(r.StatusCodes) > 0 {
		fmt.Println()
		fmt.Println("Status code distribution:")
		for code, count := range r.StatusCodes {
			fmt.Printf("  %d: %d\n", code, count)
		}
	}
	fmt.Println("======================================")
	fmt.Println()
}
