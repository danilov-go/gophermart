package memory_test

import (
	"testing"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/danilov-go/gophermart/internal/repository/memory"
	"github.com/stretchr/testify/assert"
)

const expNumber = "12345678903"

func TestMemStorage_SaveOrders(t *testing.T) {
	expAccrual := 200.5
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name     string
		number   string
		newOrder models.Order
		wantErr  error
	}{
		{
			name:   "положительный тест",
			number: expNumber,
			newOrder: models.Order{
				UserID:     7,
				Number:     expNumber,
				Status:     models.NEW,
				Accrual:    &expAccrual,
				UploadedAt: now,
			},
			wantErr: nil,
		},
		{
			name:   "номер заказа уже был загружен этим пользователем",
			number: expNumber,
			newOrder: models.Order{
				UserID:     7,
				Number:     expNumber,
				Status:     models.NEW,
				Accrual:    &expAccrual,
				UploadedAt: now,
			},
			wantErr: models.ErrOrderAlreadyUploadedBySameUser,
		},
		{
			name:   "номер заказа уже был загружен другим пользователем",
			number: expNumber,
			newOrder: models.Order{
				UserID:     5,
				Number:     expNumber,
				Status:     models.NEW,
				Accrual:    &expAccrual,
				UploadedAt: now,
			},
			wantErr: models.ErrOrderAlreadyUploadedByOtherUser,
		},
	}
	storage := memory.InitMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := storage.SaveOrders(t.Context(), tt.number, tt.newOrder)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestMemStorage_GetOrders(t *testing.T) {
	expAccrual := 200.5
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name       string
		id         int
		wantOrders []models.Orders
		wantErr    bool
	}{
		{
			name: "положительный тест",
			id:   7,
			wantOrders: []models.Orders{
				{
					Number:     expNumber,
					Status:     models.NEW,
					Accrual:    &expAccrual,
					UploadedAt: now,
				},
			},
			wantErr: false,
		},
		{
			name:       "у пользователя нет заказов",
			id:         5,
			wantOrders: nil,
			wantErr:    true,
		},
	}
	storage := memory.InitMemStorage()
	err := storage.SaveOrders(t.Context(), expNumber, models.Order{
		UserID:     7,
		Number:     expNumber,
		Status:     models.NEW,
		Accrual:    &expAccrual,
		UploadedAt: now,
	})
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders, err := storage.GetOrders(t.Context(), tt.id)
			if tt.wantErr {
				assert.ErrorIs(t, err, models.ErrNoOrdersFound)
				assert.Equal(t, tt.wantOrders, orders)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOrders, orders)
			}
		})
	}
}

func TestMemStorage_GetUnOrders(t *testing.T) {
	expAccrual := 200.5
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name       string
		wantOrders []models.Orders
		wantErr    bool
	}{
		{
			name:       "нет заказов",
			wantOrders: nil,
			wantErr:    true,
		},
		{
			name: "положительный тест",
			wantOrders: []models.Orders{
				{
					Number:     expNumber,
					Status:     models.NEW,
					Accrual:    &expAccrual,
					UploadedAt: now,
				},
				{
					Number:     expNumber + "22",
					Status:     models.PROCESSING,
					Accrual:    &expAccrual,
					UploadedAt: now,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			if !tt.wantErr {
				err := storage.SaveOrders(t.Context(), expNumber, models.Order{
					UserID:     7,
					Number:     expNumber,
					Status:     models.NEW,
					Accrual:    &expAccrual,
					UploadedAt: now,
				})
				assert.NoError(t, err)

				err = storage.SaveOrders(t.Context(), expNumber+"22", models.Order{
					UserID:     7,
					Number:     expNumber + "22",
					Status:     models.PROCESSING,
					Accrual:    &expAccrual,
					UploadedAt: now,
				})
				assert.NoError(t, err)
			}
			order, err := storage.GetUnOrders(t.Context())
			if tt.wantErr {
				assert.NoError(t, err)
				assert.Empty(t, order)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.wantOrders, order)
			}
		})
	}
}

func TestMemStorage_UpdateStatus(t *testing.T) {
	expAccrual := 200.5
	now := time.Now()
	tests := []struct {
		name    string
		accrual models.Accrual
		wantErr error
	}{
		{
			name: "положительный тест",
			accrual: models.Accrual{
				Order:   expNumber,
				Status:  "PROCESSED",
				Accrual: &expAccrual,
			},
			wantErr: nil,
		},
		{
			name: "нет заказа",
			accrual: models.Accrual{
				Order:   "unknown",
				Status:  "PROCESSED",
				Accrual: &expAccrual,
			},
			wantErr: models.ErrNoOrdersFound,
		},
	}
	storage := memory.InitMemStorage()
	err := storage.SaveOrders(t.Context(), expNumber, models.Order{
		UserID:     7,
		Number:     expNumber,
		Status:     models.NEW,
		Accrual:    &expAccrual,
		UploadedAt: now,
	})
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = storage.UpdateStatus(t.Context(), tt.accrual)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
