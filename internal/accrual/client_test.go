package accrual

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func float32Ptr(f float32) *float32 {
	return &f
}

func TestClientEvaluate(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantOrder *models.AccrualOrder
		wantErr   error
	}{
		{
			name:   "200 with valid JSON returns decoded order",
			status: http.StatusOK,
			body:   `{"order":"12345","status":"PROCESSED","accrual":500.5}`,
			wantOrder: &models.AccrualOrder{
				Order:   "12345",
				Status:  models.StatusProcessed,
				Accrual: float32Ptr(500.5),
			},
		},
		{
			name:   "200 with no accrual field returns order with nil Accrual",
			status: http.StatusOK,
			body:   `{"order":"12345","status":"REGISTERED"}`,
			wantOrder: &models.AccrualOrder{
				Order:  "12345",
				Status: models.StatusRegistered,
			},
		},
		{
			name:    "204 returns ErrAccrualClientOrderNotRegistered",
			status:  http.StatusNoContent,
			body:    "",
			wantErr: models.ErrAccrualClientOrderNotRegistered,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/orders/12345", r.URL.Path)
				w.WriteHeader(tt.status)
				if tt.body != "" {
					w.Write([]byte(tt.body))
				}
			}))
			defer srv.Close()

			client := NewClient(srv.URL, zap.NewNop())
			client.httpClient.RetryWaitMin = 10 * time.Millisecond
			client.httpClient.RetryMax = 0
			got, err := client.Evaluate("12345")

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOrder, got)
		})
	}
}

func TestClientEvaluate_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, zap.NewNop())
	client.httpClient.RetryMax = 0
	got, err := client.Evaluate("12345")

	assert.Error(t, err)
	assert.Nil(t, got)
}

func TestClientEvaluate_429ReturnsRateLimitError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, zap.NewNop())
	client.httpClient.RetryMax = 0

	got, err := client.Evaluate("12345")
	assert.Nil(t, got)
	require.Error(t, err)

	var rateLimitErr *RateLimitError
	require.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 30*time.Second, rateLimitErr.RetryAfter)
}

func TestClientEvaluate_429DefaultRetryAfter(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, zap.NewNop())
	client.httpClient.RetryMax = 0

	got, err := client.Evaluate("12345")
	assert.Nil(t, got)
	require.Error(t, err)

	var rateLimitErr *RateLimitError
	require.ErrorAs(t, err, &rateLimitErr)
	assert.Equal(t, 60*time.Second, rateLimitErr.RetryAfter)
}

func TestClientEvaluate_RetriesExhausted(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewClient(srv.URL, zap.NewNop())
	client.httpClient.RetryMax = 2
	client.httpClient.RetryWaitMin = 10 * time.Millisecond
	client.httpClient.Backoff = func(min, max time.Duration, attempt int, resp *http.Response) time.Duration {
		return min
	}

	got, err := client.Evaluate("12345")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Equal(t, int32(3), atomic.LoadInt32(&attempts))
}
