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
	"github.com/danilov-go/gophermart/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const expNumberW = "2377225624"
const expAcrual = 500.5
const expWithdraw = 200.25
const expBalance = expAcrual - expWithdraw

func TestGetBalanceHandler(t *testing.T) {

	tests := []struct {
		name  string
		login string
		code  int
		setup bool
	}{
		{
			name:  "у пользователя нет заказов",
			login: expLogin,
			code:  http.StatusOK,
			setup: false,
		},
		{
			name:  "положительный тест",
			login: expLogin,
			code:  http.StatusOK,
			setup: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.InitMemStorage()
			if tt.setup {
				err := storage.SaveOrders(expNumber, models.Order{
					Login:      expLogin,
					Status:     models.PROCESSED,
					Accrual:    expAcrual,
					UploadedAt: time.Now(),
				})
				require.NoError(t, err)
				err = storage.Withdraw(expLogin, expNumberW, expWithdraw)
				require.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			rec := httptest.NewRecorder()
			handlerFunc := handler.GetBalanceHandler()
			handlerFunc(rec, req, tt.login, storage)
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

	tests := []struct {
		name           string
		login          string
		expectedStatus int
		setup          bool
	}{
		{
			name:           "У пользователя нет истории списаний",
			login:          expLogin,
			expectedStatus: http.StatusNoContent,
			setup:          false,
		},
		{
			name:           "положительный тест",
			login:          expLogin,
			expectedStatus: http.StatusOK,
			setup:          true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := repository.InitMemStorage()
			if tt.setup {
				err := storage.Withdraw(expLogin, expNumberW, expWithdraw)
				require.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			rec := httptest.NewRecorder()
			handlerFunc := handler.GetWithdrawalsBalanceHandler()
			handlerFunc(rec, req, tt.login, storage)
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
	type orderBalance struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}
	tests := []struct {
		name           string
		login          string
		body           orderBalance
		setup          bool
		expectedStatus int
	}{
		{
			name:  "положительный тест",
			login: expLogin,
			body: orderBalance{
				Order: expNumber,
				Sum:   expWithdraw,
			},
			setup:          true,
			expectedStatus: http.StatusOK,
		},
		{
			name:  "недостаточно средств на балансе",
			login: expLogin,
			body: orderBalance{
				Order: expNumber,
				Sum:   1000.0,
			},
			setup:          true,
			expectedStatus: http.StatusPaymentRequired,
		},
		{
			name:  "неверный номер заказа по алгоритму Луна",
			login: expLogin,
			body: orderBalance{
				Order: "12345678904",
				Sum:   expAcrual,
			},
			setup:          true,
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:  "пустой номер заказа",
			login: expLogin,
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
			storage := repository.InitMemStorage()
			if tt.setup {
				err := storage.SaveOrders(expNumberW, models.Order{
					Login:      expLogin,
					Status:     models.PROCESSED,
					Accrual:    expAcrual,
					UploadedAt: time.Now(),
				})
				require.NoError(t, err)
			}
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewReader(bodyBytes))
			rec := httptest.NewRecorder()
			handlerFunc := handler.WithdrawtBalanceHandler()
			handlerFunc(rec, req, tt.login, storage)
			assert.Equal(t, tt.expectedStatus, rec.Code)
			if rec.Code == http.StatusOK {
				withdraw, err := storage.GetWithdraw(tt.login)
				assert.NoError(t, err)
				assert.NotEmpty(t, withdraw)
				assert.Equal(t, tt.body.Order, withdraw[0].Order)
				assert.Equal(t, tt.body.Sum, withdraw[0].Sum)
				balance, err := storage.GetBalance(tt.login)
				assert.NoError(t, err)
				expectedCurrent := expAcrual - tt.body.Sum
				assert.Equal(t, expectedCurrent, balance.Current)
				assert.Equal(t, tt.body.Sum, balance.Withdrawn)
			}
		})
	}
}
