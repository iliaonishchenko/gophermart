package accrual

import (
	"context"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"sync"
	"time"
)

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
}

func NewAccrualPoller(client AccrualClient, oService OrderService, bService BalanceService) *AccrualPoller {
	return &AccrualPoller{
		client:   client,
		orders:   make([]*order, 0),
		ticker:   time.NewTicker(1 * time.Second),
		oService: oService,
		bService: bService,
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
				remaining []*order
				wg        sync.WaitGroup
				mu        sync.Mutex
				sem       = make(chan struct{}, maxWorkers)
			)

			for _, o := range snapshot {
				wg.Add(1)
				sem <- struct{}{}
				go func(o *order) {
					defer wg.Done()
					defer func() { <-sem }()
					if !ap.handleOrder(ctx, o) {
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

func (ap *AccrualPoller) handleOrder(ctx context.Context, o *order) bool {
	accrualOrder, err := ap.client.Evaluate(o.number)
	if err != nil {
		logger.Log.Error("error getting order info: "+o.number, logger.Err(err))
		return false
	}
	status := string(accrualOrder.Status)
	if status == o.status {
		return false
	}

	if accrualOrder.Status == models.StatusProcessing {
		updateOrder := &models.Order{
			Number: accrualOrder.Order,
			Status: models.OrderStatusProcessing,
		}
		_, err := ap.oService.Update(ctx, updateOrder)
		if err != nil {
			logger.Log.Error("error updating order: "+o.number, logger.Err(err))
		}
		o.status = status
		return false
	}

	if accrualOrder.Status == models.StatusInvalid || accrualOrder.Status == models.StatusProcessed {
		updateOrder := &models.Order{
			Number:  o.number,
			Status:  models.OrderStatus(accrualOrder.Status),
			Accrual: accrualOrder.Accrual,
		}
		_, err := ap.oService.Update(ctx, updateOrder)
		if err != nil {
			logger.Log.Error("error updating order: "+o.number, logger.Err(err))
			return false
		}
		if accrualOrder.Status == models.StatusProcessed && accrualOrder.Accrual != nil {
			err = ap.bService.Update(ctx, accrualOrder.Accrual, o.userUUID)
			if err != nil {
				logger.Log.Error("error updating balance: "+o.number, logger.Err(err))
				return false
			}
		}
		return true
	}

	return false
}
