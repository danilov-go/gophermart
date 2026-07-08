package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danilov-go/gophermart/internal/repository/memory"

	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/danilov-go/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

var expLogin = "login1"
var expNumber = "12345678903"

func TestSaveOrderHandler(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		number      string
		id          int
		code        int
	}{
		{
			name:        "положительный тест",
			contentType: "text/plain",
			number:      expNumber,
			id:          1,
			code:        http.StatusAccepted,
		},
		{
			name:        "неверный Content-Type",
			contentType: "application/json",
			number:      expNumber,
			id:          1,
			code:        http.StatusBadRequest,
		},
		{
			name:        "пустой номер заказа",
			contentType: "text/plain",
			number:      "",
			id:          1,
			code:        http.StatusBadRequest,
		},
		{
			name:        "ошибка алгоритма Луна",
			contentType: "text/plain",
			number:      "12345678904",
			id:          1,
			code:        http.StatusUnprocessableEntity,
		},
		{
			name:        "заказ загружен этим же пользователем",
			contentType: "text/plain",
			number:      expNumber,
			id:          1,
			code:        http.StatusOK,
		},
		{
			name:        "заказ загружен другим пользователем",
			contentType: "text/plain",
			number:      expNumber,
			id:          2,
			code:        http.StatusConflict,
		},
	}
	storage := memory.InitMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.number))
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			handlerFunc := h.SaveOrderHandler()
			user := handler.AuthUser{
				ID:    tt.id,
				Login: expLogin,
			}
			handlerFunc(rec, req, user)
			assert.Equal(t, rec.Code, tt.code)
			if rec.Code == http.StatusOK || rec.Code == http.StatusAccepted {
				orders, err := storage.GetOrders(t.Context(), tt.id)
				assert.NoError(t, err)
				assert.NotEmpty(t, orders)
				var found bool
				for _, order := range orders {
					if order.Number == tt.number {
						found = true
						break
					}
				}
				assert.True(t, found)
			}

		})
	}
}

func TestGetOrderHandler(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		code      int
		checkJSON bool
	}{
		{
			name:      "у пользователя нет заказов",
			id:        1,
			code:      http.StatusNoContent,
			checkJSON: false,
		},
		{
			name:      "положительный тест",
			id:        1,
			code:      http.StatusOK,
			checkJSON: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			if tt.checkJSON {
				err := storage.SaveOrders(t.Context(), expNumber, models.Order{
					UserID:     tt.id,
					Status:     models.NEW,
					UploadedAt: time.Now(),
				})
				require.NoError(t, err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			rec := httptest.NewRecorder()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			handlerFunc := h.GetOrderHandler()
			user := handler.AuthUser{
				ID:    tt.id,
				Login: expLogin,
			}
			handlerFunc(rec, req, user)
			assert.Equal(t, tt.code, rec.Code)
			if tt.checkJSON {
				assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
				var orders []models.Orders
				err := json.Unmarshal(rec.Body.Bytes(), &orders)
				assert.NoError(t, err)
				assert.NotEmpty(t, orders)
				assert.Equal(t, expNumber, orders[0].Number)
			}
		})
	}
}
