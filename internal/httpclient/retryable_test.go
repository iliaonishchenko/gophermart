package httpclient

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRetryableClient_DoesNotRetryOn429(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewRetryableClient(zap.NewNop())
	client.RetryWaitMin = 10 * time.Millisecond
	client.RetryWaitMax = 100 * time.Millisecond

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/test", nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestRetryableClient_RetriesOn500(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := NewRetryableClient(zap.NewNop())
	client.RetryWaitMin = 10 * time.Millisecond
	client.Backoff = func(min, max time.Duration, attempt int, resp *http.Response) time.Duration {
		return min
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts))
}

func TestRetryableClient_ExhaustsRetries(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewRetryableClient(zap.NewNop())
	client.RetryMax = 2
	client.RetryWaitMin = 10 * time.Millisecond
	client.Backoff = func(min, max time.Duration, attempt int, resp *http.Response) time.Duration {
		return min
	}

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}

func TestRetryableClient_NoRetryOn4xx(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := NewRetryableClient(zap.NewNop())
	client.RetryWaitMin = 10 * time.Millisecond

	req, err := http.NewRequest(http.MethodGet, srv.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts))
}

func TestDefaultBackoff_RespectsRetryAfterHeader(t *testing.T) {
	resp := &http.Response{
		Header: http.Header{"Retry-After": []string{"5"}},
	}
	wait := DefaultBackoff(time.Second, 30*time.Second, 0, resp)
	assert.Equal(t, 5*time.Second, wait)
}

func TestDefaultBackoff_ExponentialWithCap(t *testing.T) {
	min := 1 * time.Second
	max := 10 * time.Second

	assert.Equal(t, 1*time.Second, DefaultBackoff(min, max, 0, nil))
	assert.Equal(t, 2*time.Second, DefaultBackoff(min, max, 1, nil))
	assert.Equal(t, 4*time.Second, DefaultBackoff(min, max, 2, nil))
	assert.Equal(t, 8*time.Second, DefaultBackoff(min, max, 3, nil))
	assert.Equal(t, 10*time.Second, DefaultBackoff(min, max, 4, nil))
}
