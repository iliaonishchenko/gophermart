package accrual

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/iliaonishchenko/gophermart/internal/models"
	"go.uber.org/zap"
)

var errOrderNotReady = errors.New("order not ready")

type AccrualClient interface {
	Evaluate(number string) (*models.AccrualOrder, error)
}
type OrderService interface {
	Update(ctx context.Context, order *models.Order) (*models.Order, error)
}

type BalanceService interface {
	Update(ctx context.Context, balance *float32, userUUID string) error
}

type order struct {
	number   string
	status   string
	userUUID string
}

const maxWorkers = 5

type AccrualPoller struct {
	client   AccrualClient
	orders   []*order
	mu       sync.RWMutex
	ticker   *time.Ticker
	oService OrderService
	bService BalanceService
	log      *zap.Logger
}

func NewAccrualPoller(client AccrualClient, oService OrderService, bService BalanceService, log *zap.Logger) *AccrualPoller {
	return &AccrualPoller{
		client:   client,
		orders:   make([]*order, 0),
		ticker:   time.NewTicker(1 * time.Second),
		oService: oService,
		bService: bService,
		log:      log,
	}
}

func (ap *AccrualPoller) Run(ctx context.Context) {
	for {
		select {
		case <-ap.ticker.C:
			ap.mu.RLock()
			snapshot := make([]*order, len(ap.orders))
			copy(snapshot, ap.orders)
			ap.mu.RUnlock()

			var (
				remaining    []*order
				wg           sync.WaitGroup
				mu           sync.Mutex
				sem          = make(chan struct{}, maxWorkers)
				retryAfter   time.Duration
				retryAfterMu sync.Mutex
			)

			for _, o := range snapshot {
				wg.Add(1)
				sem <- struct{}{}
				go func(o *order) {
					defer wg.Done()
					defer func() { <-sem }()

					err := ap.handleOrder(ctx, o)
					if err != nil {
						var rateLimitErr *RateLimitError
						if errors.As(err, &rateLimitErr) {
							ap.log.Warn("rate limited, pausing all workers", zap.Duration("retry_after", rateLimitErr.RetryAfter))
							retryAfterMu.Lock()
							if rateLimitErr.RetryAfter > retryAfter {
								retryAfter = rateLimitErr.RetryAfter
							}
							retryAfterMu.Unlock()
						} else if !errors.Is(err, errOrderNotReady) {
							ap.log.Error("error handling order: "+o.number, zap.Error(err))
						}
						mu.Lock()
						remaining = append(remaining, o)
						mu.Unlock()
					}
				}(o)
			}
			wg.Wait()

			ap.mu.Lock()
			ap.orders = remaining
			ap.mu.Unlock()

			if retryAfter > 0 {
				select {
				case <-time.After(retryAfter):
				case <-ctx.Done():
					return
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

func (ap *AccrualPoller) Add(number string, status string, userUUID string) {
	ap.mu.Lock()
	defer ap.mu.Unlock()
	ap.orders = append(ap.orders, &order{number: number, status: status, userUUID: userUUID})
}

func (ap *AccrualPoller) handleOrder(ctx context.Context, o *order) error {
	accrualOrder, err := ap.client.Evaluate(o.number)
	if err != nil {
		return fmt.Errorf("getting order info: %w", err)
	}
	status := string(accrualOrder.Status)
	if status == o.status {
		return errOrderNotReady
	}

	if accrualOrder.Status == models.StatusProcessing {
		updateOrder := &models.Order{
			Number: accrualOrder.Order,
			Status: models.OrderStatusProcessing,
		}
		_, err := ap.oService.Update(ctx, updateOrder)
		if err != nil {
			return fmt.Errorf("updating order status to processing: %w", err)
		}
		o.status = status
		return errOrderNotReady
	}

	if accrualOrder.Status == models.StatusInvalid || accrualOrder.Status == models.StatusProcessed {
		updateOrder := &models.Order{
			Number:  o.number,
			Status:  models.OrderStatus(accrualOrder.Status),
			Accrual: accrualOrder.Accrual,
		}
		_, err := ap.oService.Update(ctx, updateOrder)
		if err != nil {
			return fmt.Errorf("updating order: %w", err)
		}
		if accrualOrder.Status == models.StatusProcessed && accrualOrder.Accrual != nil {
			err = ap.bService.Update(ctx, accrualOrder.Accrual, o.userUUID)
			if err != nil {
				return fmt.Errorf("updating balance: %w", err)
			}
		}
		return nil
	}

	return errOrderNotReady
}