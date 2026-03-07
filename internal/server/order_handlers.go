package server

import (
	"context"
	"errors"
	"github.com/iliaonishchenko/gophermart/internal/auth"
	"github.com/iliaonishchenko/gophermart/internal/logger"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	"strings"
)

func (s *Server) PostAPIUserOrders(ctx context.Context, request api.PostAPIUserOrdersRequestObject) (api.PostAPIUserOrdersResponseObject, error) {
	userUUID, ok := auth.GetUserUUID(ctx)
	if !ok {
		return api.PostAPIUserOrders400Response{}, nil
	}
	orderNumber := strings.TrimSpace(*request.Body)

	newOrder := &models.Order{
		Number:  orderNumber,
		UserID:  userUUID,
		Status:  models.OrderStatusNew,
		Accrual: nil,
	}
	createdOrder, err := s.orderService.Create(ctx, newOrder)
	if errors.Is(err, models.ErrInvalidOrderFormat) {
		logger.Log.Error("invalid order", logger.Err(err))
		return api.PostAPIUserOrders422Response{}, nil
	}
	if errors.Is(err, models.ErrOrderAlreadyExists) {
		logger.Log.Info("order already exists", logger.Err(err))
		return api.PostAPIUserOrders200Response{}, nil
	}
	if errors.Is(err, models.ErrOrderExistsDifferentUser) {
		logger.Log.Error("order already uploaded by another user", logger.Err(err))
		return api.PostAPIUserOrders409Response{}, nil
	}
	if err != nil {
		logger.Log.Error("could not create order", logger.Err(err))
		return api.PostAPIUserOrders500Response{}, nil
	}
	s.accrualService.Add(createdOrder.Number, string(models.StatusRegistered), createdOrder.UserID)
	return api.PostAPIUserOrders202Response{}, nil
}

func (s *Server) GetAPIUserOrders(ctx context.Context, request api.GetAPIUserOrdersRequestObject) (api.GetAPIUserOrdersResponseObject, error) {
	userUUID := ctx.Value(auth.UserUUIDKey).(string)
	orders, err := s.orderService.Get(ctx, &userUUID)
	if err != nil {
		logger.Log.Error("could not get order", logger.Err(err))
		return api.GetAPIUserOrders500Response{}, nil
	}
	if len(orders) == 0 {
		return api.GetAPIUserOrders204Response{}, nil
	}

	return api.GetAPIUserOrders200JSONResponse(s.orderToAPIOrder(orders)), nil
}

func (s *Server) orderToAPIOrder(orders []*models.Order) []api.Order {
	results := make([]api.Order, 0, len(orders))
	for _, order := range orders {
		apiOrder := api.Order{
			Accrual:    order.Accrual,
			Number:     order.Number,
			Status:     api.OrderStatus(order.Status),
			UploadedAt: order.UploadedAt,
		}

		results = append(results, apiOrder)
	}
	return results
}
