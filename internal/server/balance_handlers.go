package server

import (
	"context"
	"github.com/iliaonishchenko/gophermart/pkg/api"
)

func (s *Server) GetAPIUserBalance(ctx context.Context, r api.GetAPIUserBalanceRequestObject) (api.GetAPIUserBalanceResponseObject, error) {
	balance, err := s.balanceService.Get(ctx)
	if err != nil {
		return api.GetAPIUserBalance500Response{}, nil
	}

	return api.GetAPIUserBalance200JSONResponse{
		Current:   balance.Current,
		Withdrawn: balance.Withdrawn,
	}, nil
}
