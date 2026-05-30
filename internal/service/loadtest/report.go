package loadtest

import "time"

func buildReport(results []result, duration time.Duration) Report {
	report := Report{
		TotalRequests: len(results),
		TotalDuration: duration,
		StatusCodes:   make(map[int]int),
	}

	for _, r := range results {
		if r.err != nil {
			report.ErrorCount++
			continue
		}

		report.StatusCodes[r.statusCode]++

		if r.statusCode == 200 {
			report.SuccessCount++
		}
	}

	return report
}
