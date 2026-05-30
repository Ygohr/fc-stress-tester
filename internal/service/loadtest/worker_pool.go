package loadtest

import (
	"context"
	"net/http"
	"sync"

	"github.com/Ygohr/fc-stress-tester/internal/contract"
)

type workerPool struct {
	client      contract.HTTPClient
	concurrency int
}

func newWorkerPool(client contract.HTTPClient, concurrency int) *workerPool {
	return &workerPool{
		client:      client,
		concurrency: concurrency,
	}
}

func (wp *workerPool) run(ctx context.Context, url string, totalRequests int) []result {
	jobs := make(chan struct{}, totalRequests)
	results := make(chan result, totalRequests)

	for i := 0; i < totalRequests; i++ {
		jobs <- struct{}{}
	}
	close(jobs)

	var wg sync.WaitGroup
	for i := 0; i < wp.concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			wp.work(ctx, url, jobs, results)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	collected := make([]result, 0, totalRequests)
	for r := range results {
		collected = append(collected, r)
	}

	return collected
}

func (wp *workerPool) work(ctx context.Context, url string, jobs <-chan struct{}, results chan<- result) {
	for range jobs {
		select {
		case <-ctx.Done():
			results <- result{err: ctx.Err()}
			continue
		default:
		}

		results <- wp.doRequest(ctx, url)
	}
}

func (wp *workerPool) doRequest(ctx context.Context, url string) result {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return result{err: err}
	}

	resp, err := wp.client.Do(req)
	if err != nil {
		return result{err: err}
	}
	defer resp.Body.Close()

	return result{statusCode: resp.StatusCode}
}
