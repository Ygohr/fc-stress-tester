package loadtest

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

type concurrencyTrackingClient struct {
	mu          sync.Mutex
	active      int64
	maxObserved int64
}

func (c *concurrencyTrackingClient) Do(req *http.Request) (*http.Response, error) {
	current := atomic.AddInt64(&c.active, 1)
	defer atomic.AddInt64(&c.active, -1)

	c.mu.Lock()
	if current > c.maxObserved {
		c.maxObserved = current
	}
	c.mu.Unlock()

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}, nil
}

func TestWorkerPool_ProcessesAllJobs(t *testing.T) {
	client := &concurrencyTrackingClient{}
	pool := newWorkerPool(client, 5)

	results := pool.run(context.Background(), "http://example.com", 30)

	assert.Len(t, results, 30)
}

func TestWorkerPool_ProcessesExactNumberOfRequests(t *testing.T) {
	client := &concurrencyTrackingClient{}
	pool := newWorkerPool(client, 3)

	results := pool.run(context.Background(), "http://example.com", 17)

	assert.Len(t, results, 17)
}

func TestWorkerPool_RespectsConfiguredConcurrency(t *testing.T) {
	client := &concurrencyTrackingClient{}
	concurrency := 4
	pool := newWorkerPool(client, concurrency)

	pool.run(context.Background(), "http://example.com", 50)

	assert.LessOrEqual(t, client.maxObserved, int64(concurrency),
		"max concurrent workers (%d) must not exceed concurrency (%d)",
		client.maxObserved, concurrency)
}

func TestWorkerPool_TerminatesCorrectly(t *testing.T) {
	client := &concurrencyTrackingClient{}
	pool := newWorkerPool(client, 5)

	done := make(chan struct{})
	go func() {
		pool.run(context.Background(), "http://example.com", 20)
		close(done)
	}()

	select {
	case <-done:
	}
}

func TestWorkerPool_CancelledContextStopsWorkers(t *testing.T) {
	client := &concurrencyTrackingClient{}
	pool := newWorkerPool(client, 5)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results := pool.run(ctx, "http://example.com", 50)

	assert.Len(t, results, 50)
	for _, r := range results {
		assert.Error(t, r.err)
	}
}
