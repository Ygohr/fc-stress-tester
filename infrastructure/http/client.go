package http

import (
	"net/http"
	"time"
)

const defaultTimeout = 30 * time.Second

type HTTPClient struct {
	client *http.Client
}

func NewHTTPClient() *HTTPClient {
	return NewHTTPClientWithTimeout(defaultTimeout)
}

func NewHTTPClientWithTimeout(timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *HTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}
