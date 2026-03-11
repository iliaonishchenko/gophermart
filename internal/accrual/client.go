package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/iliaonishchenko/gophermart/internal/httpclient"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"go.uber.org/zap"
)

type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %s", e.RetryAfter)
}

type Client struct {
	httpClient *httpclient.RetryableClient
	addr       string
	log        *zap.Logger
}

func NewClient(addr string, log *zap.Logger) *Client {
	return &Client{
		httpClient: httpclient.NewRetryableClient(log),
		addr:       addr,
		log:        log,
	}
}

func (c *Client) Evaluate(number string) (*models.AccrualOrder, error) {
	uri := c.addr + "/api/orders/" + number
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		retryAfter := 60 * time.Second
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if seconds, err := strconv.ParseInt(ra, 10, 64); err == nil {
				retryAfter = time.Duration(seconds) * time.Second
			}
		}
		return nil, &RateLimitError{RetryAfter: retryAfter}
	}

	if resp.StatusCode == http.StatusNoContent {
		c.log.Warn("заказ не зарегистрирован в системе расчёта")
		return nil, models.ErrAccrualClientOrderNotRegistered
	} else if resp.StatusCode != http.StatusOK {
		c.log.Warn("unexpected response status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var accrualOrder models.AccrualOrder
	if err := json.NewDecoder(resp.Body).Decode(&accrualOrder); err != nil {
		return nil, err
	}
	return &accrualOrder, nil
}
