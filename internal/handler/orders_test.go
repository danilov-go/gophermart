package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mem "github.com/danilov-go/gophermart/internal/repository/mem_storage"

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
		body        string
		login       string
		code        int
	}{
		{
			name:        "положительный тест",
			contentType: "text/plain",
			body:        expNumber,
			login:       expLogin,
			code:        http.StatusAccepted,
		},
		{
			name:        "неверный Content-Type",
			contentType: "application/json",
			body:        expNumber,
			login:       expLogin,
			code:        http.StatusBadRequest,
		},
		{
			name:        "пустой номер заказа",
			contentType: "text/plain",
			body:        "",
			login:       expLogin,
			code:        http.StatusBadRequest,
		},
		{
			name:        "ошибка алгоритма Луна",
			contentType: "text/plain",
			body:        "12345678904",
			login:       expLogin,
			code:        http.StatusUnprocessableEntity,
		},
		{
			name:        "заказ загружен этим же пользователем",
			contentType: "text/plain",
			body:        expNumber,
			login:       expLogin,
			code:        http.StatusOK,
		},
		{
			name:        "заказ загружен другим пользователем",
			contentType: "text/plain",
			body:        expNumber,
			login:       "login2",
			code:        http.StatusConflict,
		},
	}
	storage := mem.InitMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			rec := httptest.NewRecorder()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			handlerFunc := h.SaveOrderHandler()
			handlerFunc(rec, req, tt.login)
			assert.Equal(t, rec.Code, tt.code)
			if rec.Code == http.StatusOK || rec.Code == http.StatusAccepted {
				orders, err := storage.GetOrders(tt.login)
				assert.NoError(t, err)
				assert.NotEmpty(t, orders)
				var found bool
				for _, order := range orders {
					if order.Number == tt.body {
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
		login     string
		code      int
		checkJSON bool
	}{
		{
			name:      "у пользователя нет заказов",
			login:     expLogin,
			code:      http.StatusNoContent,
			checkJSON: false,
		},
		{
			name:      "положительный тест",
			login:     expLogin,
			code:      http.StatusOK,
			checkJSON: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := mem.InitMemStorage()
			if tt.checkJSON {
				err := storage.SaveOrders(expNumber, models.Order{
					Login:      expLogin,
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
			handlerFunc(rec, req, tt.login)
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
