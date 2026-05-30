package loadtest

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Ygohr/fc-stress-tester/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func okResponse() *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

func responseWithStatus(code int) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       io.NopCloser(strings.NewReader("")),
	}
}

type flexMockClient struct {
	responses []*http.Response
	idx       int
}

func (f *flexMockClient) Do(req *http.Request) (*http.Response, error) {
	resp := f.responses[f.idx%len(f.responses)]
	f.idx++
	return resp, nil
}

func TestService_ExecutesExactNumberOfRequests(t *testing.T) {
	mockClient := &testutil.MockHTTPClient{}
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(okResponse(), nil)

	svc := NewService(mockClient)
	cfg := Config{URL: "http://example.com", Requests: 50, Concurrency: 5}

	report, err := svc.Execute(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 50, report.TotalRequests)
	mockClient.AssertNumberOfCalls(t, "Do", 50)
}

func TestService_CountsHTTP200Responses(t *testing.T) {
	mockClient := &testutil.MockHTTPClient{}
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(okResponse(), nil)

	svc := NewService(mockClient)
	cfg := Config{URL: "http://example.com", Requests: 10, Concurrency: 2}

	report, err := svc.Execute(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 10, report.SuccessCount)
	assert.Equal(t, 0, report.ErrorCount)
}

func TestService_AggregatesStatusCodes(t *testing.T) {
	responses := []*http.Response{
		responseWithStatus(200),
		responseWithStatus(200),
		responseWithStatus(404),
		responseWithStatus(200),
		responseWithStatus(404),
		responseWithStatus(200),
	}

	flexClient := &flexMockClient{responses: responses}
	svc := NewService(flexClient)
	cfg := Config{URL: "http://example.com", Requests: 6, Concurrency: 1}

	report, err := svc.Execute(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 6, report.TotalRequests)
	assert.Equal(t, 4, report.StatusCodes[200])
	assert.Equal(t, 2, report.StatusCodes[404])
}

func TestService_CountsRequestErrors(t *testing.T) {
	mockClient := &testutil.MockHTTPClient{}
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).
		Return((*http.Response)(nil), assert.AnError)

	svc := NewService(mockClient)
	cfg := Config{URL: "http://example.com", Requests: 5, Concurrency: 1}

	report, err := svc.Execute(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 5, report.ErrorCount)
	assert.Equal(t, 0, report.SuccessCount)
}

func TestService_CalculatesExecutionReport(t *testing.T) {
	mockClient := &testutil.MockHTTPClient{}
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(okResponse(), nil)

	svc := NewService(mockClient)
	cfg := Config{URL: "http://example.com", Requests: 10, Concurrency: 2}

	report, err := svc.Execute(context.Background(), cfg)

	assert.NoError(t, err)
	assert.Equal(t, 10, report.TotalRequests)
	assert.GreaterOrEqual(t, report.TotalDuration.Nanoseconds(), int64(0))
	assert.NotNil(t, report.StatusCodes)
}

func TestService_SupportsContextCancellation(t *testing.T) {
	mockClient := &testutil.MockHTTPClient{}
	mockClient.On("Do", mock.AnythingOfType("*http.Request")).Return(okResponse(), nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	svc := NewService(mockClient)
	cfg := Config{URL: "http://example.com", Requests: 100, Concurrency: 5}

	report, err := svc.Execute(ctx, cfg)

	assert.NoError(t, err)
	assert.Equal(t, 100, report.TotalRequests)
	assert.Equal(t, 100, report.ErrorCount)
}
