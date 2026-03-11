package httpclient

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type CheckRetry func(resp *http.Response, err error) bool

type Backoff func(min, max time.Duration, attempt int, resp *http.Response) time.Duration

type RetryableClient struct {
	HTTPClient   *http.Client
	RetryMax     int
	RetryWaitMin time.Duration
	RetryWaitMax time.Duration
	CheckRetry   CheckRetry
	Backoff      Backoff
	Logger       *zap.Logger
}

func NewRetryableClient(logger *zap.Logger) *RetryableClient {
	return &RetryableClient{
		HTTPClient:   &http.Client{},
		RetryMax:     3,
		RetryWaitMin: 1 * time.Second,
		RetryWaitMax: 30 * time.Second,
		CheckRetry:   DefaultCheckRetry,
		Backoff:      DefaultBackoff,
		Logger:       logger,
	}
}

func (c *RetryableClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; ; attempt++ {
		resp, err = c.HTTPClient.Do(req)

		shouldRetry := c.CheckRetry(resp, err)
		if !shouldRetry || attempt >= c.RetryMax {
			break
		}

		if resp != nil {
			resp.Body.Close()
		}

		wait := c.Backoff(c.RetryWaitMin, c.RetryWaitMax, attempt, resp)
		c.Logger.Warn(
			fmt.Sprintf("retrying request (attempt %d/%d)", attempt+1, c.RetryMax),
			zap.String("url", req.URL.String()),
			zap.Duration("backoff", wait),
		)

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(wait):
		}
	}

	return resp, err
}

func DefaultCheckRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp.StatusCode == 0 || (resp.StatusCode >= 500 && resp.StatusCode != http.StatusNotImplemented) {
		return true
	}
	return false
}

func DefaultBackoff(min, max time.Duration, attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
			if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
				return time.Duration(seconds) * time.Second
			}
		}
	}

	wait := time.Duration(math.Pow(2, float64(attempt))) * min
	if wait > max {
		wait = max
	}
	return wait
}
