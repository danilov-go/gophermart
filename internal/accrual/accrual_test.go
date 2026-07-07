package accrual_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danilov-go/gophermart/internal/repository/memory"

	"github.com/danilov-go/gophermart/internal/accrual"
	"github.com/danilov-go/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const (
	expId      = 1
	expOrder   = "12345678903"
	expStatus  = "NEW"
	expAccrual = 0.0
)

func TestWorker_TableDriven(t *testing.T) {
	expAccrual := 500.5
	zeroAccrual := 0.0
	tests := []struct {
		name    string
		code    int
		accrual models.Accrual
		setup   bool
	}{
		{
			name: "положительный тест",
			code: http.StatusOK,
			accrual: models.Accrual{
				Order:   expOrder,
				Status:  "PROCESSED",
				Accrual: &expAccrual,
			},
			setup: true,
		},
		{
			name: "заказ отклонен",
			code: http.StatusOK,
			accrual: models.Accrual{
				Order:   "2377225624",
				Status:  "INVALID",
				Accrual: &zeroAccrual,
			},
			setup: true,
		},
		{
			name:    "внутренняя ошибка сервера",
			code:    http.StatusInternalServerError,
			accrual: models.Accrual{},
			setup:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.setup {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.code)
					err := json.NewEncoder(w).Encode(tt.accrual)
					require.NoError(t, err)
					return
				}
				w.WriteHeader(tt.code)
			}))
			defer server.Close()
			storage := memory.InitMemStorage()
			logger := zaptest.NewLogger(t)
			order := expOrder
			if tt.setup {
				order = tt.accrual.Order
			}
			err := storage.SaveOrders(t.Context(), order, models.Order{
				UserID:     1,
				Status:     models.NEW,
				UploadedAt: time.Now(),
			})
			require.NoError(t, err)
			agent := accrual.New(1, server.URL, logger.Sugar(), storage)
			ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
			defer cancel()
			go agent.Worker(ctx)
			<-ctx.Done()
			storageOrders, err := storage.GetOrders(t.Context(), expId)
			require.NoError(t, err)
			require.NotEmpty(t, storageOrders)
			expectideStatus := expStatus
			expectideAccrual := expAccrual
			if tt.setup {
				expectideStatus = tt.accrual.Status
				expectideAccrual = *tt.accrual.Accrual
			}
			if storageOrders[0].Accrual != nil {
				actualAccrual := *storageOrders[0].Accrual
				assert.Equal(t, expectideAccrual, actualAccrual)
			}
			assert.Equal(t, expectideStatus, storageOrders[0].Status)

		})
	}
}
