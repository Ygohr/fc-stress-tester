package loadtest

import "time"

type Config struct {
	URL         string
	Requests    int
	Concurrency int
}

type Report struct {
	TotalRequests int
	TotalDuration time.Duration
	SuccessCount  int
	ErrorCount    int
	StatusCodes   map[int]int
}

type result struct {
	statusCode int
	err        error
}
