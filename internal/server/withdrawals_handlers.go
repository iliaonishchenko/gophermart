package server

import (
	"context"
	"errors"
	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/iliaonishchenko/gophermart/pkg/api"
	"time"
)

func (s *Server) PostAPIUserBalanceWithdraw(ctx context.Context, r api.PostAPIUserBalanceWithdrawRequestObject) (api.PostAPIUserBalanceWithdrawResponseObject, error) {
	withdrawalToCreate := &models.Withdrawal{
		Order: r.Body.Order,
		Sum:   r.Body.Sum,
	}
	_, err := s.withdrawalService.Create(ctx, withdrawalToCreate)
	if errors.Is(err, models.ErrWithdrawalNotEnoughFunds) {
		return api.PostAPIUserBalanceWithdraw402Response{}, nil
	}
	if errors.Is(err, models.ErrWithdrawalNonExistentOrder) {
		return api.PostAPIUserBalanceWithdraw422Response{}, nil
	}
	if err != nil {
		return api.PostAPIUserBalanceWithdraw500Response{}, nil
	}

	return api.PostAPIUserBalanceWithdraw200Response{}, nil
}

func (s *Server) GetAPIUserWithdrawals(ctx context.Context, r api.GetAPIUserWithdrawalsRequestObject) (api.GetAPIUserWithdrawalsResponseObject, error) {
	withdrawals, err := s.withdrawalService.Get(ctx)
	if err != nil {
		return api.GetAPIUserWithdrawals500Response{}, nil
	}
	if len(withdrawals) == 0 {
		return api.GetAPIUserWithdrawals204Response{}, nil
	}

	return api.GetAPIUserWithdrawals200JSONResponse(s.withdrawalToAPIWithdrawal(withdrawals)), nil
}

func (s *Server) withdrawalToAPIWithdrawal(withdrawals []*models.Withdrawal) []api.WithdrawalResponse {
	apiWithdrawals := make([]api.WithdrawalResponse, 0, len(withdrawals))
	for _, w := range withdrawals {
		apiWithdrawal := api.WithdrawalResponse{
			Order:       w.Order,
			ProcessedAt: w.ProcessedAt.Format(time.RFC3339),
			Sum:         w.Sum,
		}
		apiWithdrawals = append(apiWithdrawals, apiWithdrawal)
	}
	return apiWithdrawals
}
