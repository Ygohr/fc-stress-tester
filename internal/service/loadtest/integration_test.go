package loadtest_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	httpinfra "github.com/Ygohr/fc-stress-tester/infrastructure/http"
	"github.com/Ygohr/fc-stress-tester/internal/service/loadtest"
	"github.com/Ygohr/fc-stress-tester/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_SuccessfulRequests(t *testing.T) {
	srv := testutil.NewTestServer(http.StatusOK)
	defer srv.Close()

	client := httpinfra.NewHTTPClient()
	svc := loadtest.NewService(client)

	report, err := svc.Execute(context.Background(), loadtest.Config{
		URL:         srv.URL,
		Requests:    20,
		Concurrency: 5,
	})

	require.NoError(t, err)
	assert.Equal(t, 20, report.TotalRequests)
	assert.Equal(t, 20, report.SuccessCount)
	assert.Equal(t, 0, report.ErrorCount)
	assert.Equal(t, 20, report.StatusCodes[200])
}

func TestIntegration_ReportGeneration(t *testing.T) {
	srv := testutil.NewTestServer(http.StatusOK)
	defer srv.Close()

	client := httpinfra.NewHTTPClient()
	svc := loadtest.NewService(client)

	report, err := svc.Execute(context.Background(), loadtest.Config{
		URL:         srv.URL,
		Requests:    10,
		Concurrency: 2,
	})

	require.NoError(t, err)
	assert.Greater(t, report.TotalDuration.Nanoseconds(), int64(0))
	assert.NotNil(t, report.StatusCodes)
}

func TestIntegration_StatusCodeAggregation(t *testing.T) {
	var count int64
	srv := testutil.NewTestServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt64(&count, 1)
		if n%2 == 0 {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})
	defer srv.Close()

	client := httpinfra.NewHTTPClient()
	svc := loadtest.NewService(client)

	report, err := svc.Execute(context.Background(), loadtest.Config{
		URL:         srv.URL,
		Requests:    10,
		Concurrency: 1,
	})

	require.NoError(t, err)
	assert.Equal(t, 10, report.TotalRequests)
	assert.Equal(t, 5, report.StatusCodes[200])
	assert.Equal(t, 5, report.StatusCodes[404])
}

func TestIntegration_ConcurrencyExecution(t *testing.T) {
	var active, maxActive int64
	var mu int64
	_ = mu

	srv := testutil.NewTestServerWithHandler(func(w http.ResponseWriter, r *http.Request) {
		current := atomic.AddInt64(&active, 1)
		defer atomic.AddInt64(&active, -1)
		for {
			old := atomic.LoadInt64(&maxActive)
			if current <= old || atomic.CompareAndSwapInt64(&maxActive, old, current) {
				break
			}
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	client := httpinfra.NewHTTPClient()
	svc := loadtest.NewService(client)

	concurrency := 5
	_, err := svc.Execute(context.Background(), loadtest.Config{
		URL:         srv.URL,
		Requests:    30,
		Concurrency: concurrency,
	})

	require.NoError(t, err)
	assert.LessOrEqual(t, maxActive, int64(concurrency))
}

func TestIntegration_ErrorCounting(t *testing.T) {
	client := httpinfra.NewHTTPClient()
	svc := loadtest.NewService(client)

	report, err := svc.Execute(context.Background(), loadtest.Config{
		URL:         "http://127.0.0.1:1",
		Requests:    5,
		Concurrency: 2,
	})

	require.NoError(t, err)
	assert.Equal(t, 5, report.TotalRequests)
	assert.Equal(t, 5, report.ErrorCount)
	assert.Equal(t, 0, report.SuccessCount)
}
