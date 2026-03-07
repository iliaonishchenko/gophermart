package accrual

import (
	"encoding/json"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"net/http"
)

type Client struct {
	http.Client
	addr string
}

func NewClient(addr string) *Client {
	return &Client{addr: addr}
}

func (c *Client) Evaluate(number string) (*models.AccrualOrder, error) {
	uri := c.addr + "/api/orders/" + number
	resp, err := c.Get(uri)
	if err != nil {
		logger.Log.Error("fails to evaluate accrual", logger.Err(err))
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNoContent {
		logger.Log.Warn("заказ не зарегистрирован в системе расчёта", logger.Err(err))
		return nil, models.ErrAccrualClientOrderNotRegistered
	} else if resp.StatusCode == http.StatusTooManyRequests {
		logger.Log.Warn("превышено количество запросов к сервису", logger.Err(err))
		return nil, models.ErrAccrualClientTooManyRequests
	} else if resp.StatusCode != http.StatusOK {
		logger.Log.Warn("внутренняя ошибка сервера", logger.Err(err))
		return nil, models.ErrAccrualClientInternalError
	}

	var accrualOrder models.AccrualOrder
	if err := json.NewDecoder(resp.Body).Decode(&accrualOrder); err != nil {
		logger.Log.Error("fails to decode accrual order", logger.Err(err))
		return nil, err
	}
	return &accrualOrder, nil
}
