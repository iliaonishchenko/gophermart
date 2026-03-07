package accrual

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/iliaonishchenko/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func float32Ptr(f float32) *float32 {
	return &f
}

func TestClientEvaluate(t *testing.T) {
	tests := []struct {
		name        string
		status      int
		body        string
		wantOrder   *models.AccrualOrder
		wantErr     error
		wantAnyErr  bool
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
		{
			name:    "429 returns ErrAccrualClientTooManyRequests",
			status:  http.StatusTooManyRequests,
			body:    "",
			wantErr: models.ErrAccrualClientTooManyRequests,
		},
		{
			name:    "500 returns ErrAccrualClientInternalError",
			status:  http.StatusInternalServerError,
			body:    "",
			wantErr: models.ErrAccrualClientInternalError,
		},
		{
			name:       "200 with malformed JSON returns decode error",
			status:     http.StatusOK,
			body:       `{invalid json`,
			wantAnyErr: true,
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

			client := NewClient(srv.URL)
			got, err := client.Evaluate("12345")

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, got)
				return
			}
			if tt.wantAnyErr {
				assert.Error(t, err)
				assert.Nil(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantOrder, got)
		})
	}
}
