package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/danilov-go/gophermart/internal/models"
	"github.com/danilov-go/gophermart/internal/repository/memory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

const expNumberW = "2377225624"
const expWithdraw = 200.25

func TestGetBalanceHandler(t *testing.T) {
	expAcrual := 500.5
	expBalance := expAcrual - expWithdraw
	tests := []struct {
		name  string
		id    int
		code  int
		setup bool
	}{
		{
			name:  "у пользователя нет заказов",
			id:    1,
			code:  http.StatusOK,
			setup: false,
		},
		{
			name:  "положительный тест",
			id:    1,
			code:  http.StatusOK,
			setup: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			if tt.setup {
				err := storage.SaveOrders(t.Context(), expNumber, models.Order{
					UserID:     1,
					Status:     models.PROCESSED,
					Accrual:    &expAcrual,
					UploadedAt: time.Now(),
				})
				require.NoError(t, err)
				err = storage.Withdraw(t.Context(), tt.id, expNumberW, expWithdraw)
				require.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			rec := httptest.NewRecorder()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			handlerFunc := h.GetBalanceHandler()
			user := handler.AuthUser{
				ID:    tt.id,
				Login: expLogin,
			}
			handlerFunc(rec, req, user)
			assert.Equal(t, tt.code, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			var balance models.Balance
			err := json.Unmarshal(rec.Body.Bytes(), &balance)
			assert.NoError(t, err)
			if tt.setup {
				assert.Equal(t, expBalance, balance.Current)
				assert.Equal(t, expWithdraw, balance.Withdrawn)
			} else {
				assert.Equal(t, 0.0, balance.Current)
				assert.Equal(t, 0.0, balance.Withdrawn)
			}
		})
	}
}

func TestGetWithdrawalsBalanceHandler(t *testing.T) {
	expAcrual := 500.5
	tests := []struct {
		name           string
		id             int
		expectedStatus int
		setup          bool
	}{
		{
			name:           "У пользователя нет истории списаний",
			id:             1,
			expectedStatus: http.StatusNoContent,
			setup:          false,
		},
		{
			name:           "положительный тест",
			id:             1,
			expectedStatus: http.StatusOK,
			setup:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			if tt.setup {
				err := storage.SaveOrders(t.Context(), expNumber, models.Order{
					UserID:     tt.id,
					Status:     models.PROCESSED,
					Accrual:    &expAcrual,
					UploadedAt: time.Now(),
				})
				require.NoError(t, err)
				err = storage.Withdraw(t.Context(), tt.id, expNumberW, expWithdraw)
				require.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			rec := httptest.NewRecorder()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			handlerFunc := h.GetWithdrawalsBalanceHandler()
			user := handler.AuthUser{
				ID:    tt.id,
				Login: expLogin,
			}
			handlerFunc(rec, req, user)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if tt.setup {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
				var withdrawals []models.Withdraw
				err := json.Unmarshal(rec.Body.Bytes(), &withdrawals)
				assert.NoError(t, err)
				require.NotEmpty(t, withdrawals)
				assert.Equal(t, expNumberW, withdrawals[0].Order)
				assert.Equal(t, expWithdraw, withdrawals[0].Sum)
			} else {
				assert.Empty(t, rec.Body.Bytes())
			}
		})
	}
}

func TestWithdrawtBalanceHandler(t *testing.T) {
	expAcrual := 500.5
	type orderBalance struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	tests := []struct {
		name           string
		id             int
		body           orderBalance
		setup          bool
		expectedStatus int
	}{
		{
			name: "положительный тест",
			id:   1,
			body: orderBalance{
				Order: expNumber,
				Sum:   expWithdraw,
			},
			setup:          true,
			expectedStatus: http.StatusOK,
		},
		{
			name: "недостаточно средств на балансе",
			id:   1,
			body: orderBalance{
				Order: expNumber,
				Sum:   1000.0,
			},
			setup:          true,
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name: "неверный номер заказа по алгоритму Луна",
			id:   1,
			body: orderBalance{
				Order: "12345678904",
				Sum:   expAcrual,
			},
			setup:          true,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "пустой номер заказа",
			id:   1,
			body: orderBalance{
				Order: "",
				Sum:   expAcrual,
			},
			setup:          false,
			expectedStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			if tt.setup {
				err := storage.SaveOrders(t.Context(), expNumberW, models.Order{
					UserID:     tt.id,
					Status:     models.PROCESSED,
					Accrual:    &expAcrual,
					UploadedAt: time.Now(),
				})
				require.NoError(t, err)
			}
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			handlerFunc := h.WithdrawtBalanceHandler()
			user := handler.AuthUser{
				ID:    tt.id,
				Login: expLogin,
			}
			handlerFunc(rec, req, user)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if rec.Code == http.StatusOK {
				withdraw, err := storage.GetWithdraw(t.Context(), tt.id)
				assert.NoError(t, err)
				assert.NotEmpty(t, withdraw)
				assert.Equal(t, tt.body.Order, withdraw[0].Order)
				assert.Equal(t, tt.body.Sum, withdraw[0].Sum)
				balance, err := storage.GetBalance(t.Context(), tt.id)
				assert.NoError(t, err)
				expectedCurrent := expAcrual - tt.body.Sum
				assert.Equal(t, expectedCurrent, balance.Current)
				assert.Equal(t, tt.body.Sum, balance.Withdrawn)
			}
		})
	}
}
