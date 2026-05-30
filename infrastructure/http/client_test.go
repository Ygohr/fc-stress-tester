package http_test

import (
	"net/http"
	"testing"
	"time"

	httpinfra "github.com/Ygohr/fc-stress-tester/infrastructure/http"
	"github.com/Ygohr/fc-stress-tester/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPClient_ConfiguresDefaultTimeout(t *testing.T) {
	client := httpinfra.NewHTTPClient()
	assert.NotNil(t, client)
}

func TestHTTPClient_ConfiguresCustomTimeout(t *testing.T) {
	client := httpinfra.NewHTTPClientWithTimeout(5 * time.Second)
	assert.NotNil(t, client)
}

func TestHTTPClient_CreatesValidRequests(t *testing.T) {
	srv := testutil.NewTestServer(http.StatusOK)
	defer srv.Close()

	client := httpinfra.NewHTTPClient()

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
