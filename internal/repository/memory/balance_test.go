package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/danilov-go/gophermart/internal/repository/memory"
	"github.com/stretchr/testify/assert"
)

const expBal = 200.5

func TestMemStorage_GetBalance(t *testing.T) {
	expAccrual := 500.5
	now := time.Now()
	tests := []struct {
		name        string
		id          int
		wantBalance models.Balance
	}{
		{
			name: "положительный тест",
			id:   7,
			wantBalance: models.Balance{
				Current:   300.0,
				Withdrawn: 200.5,
			},
		},
		{
			name: "нет заказов",
			id:   5,
			wantBalance: models.Balance{
				Current:   0.0,
				Withdrawn: 0.0,
			},
		},
	}
	storage := memory.InitMemStorage()
	err := storage.SaveOrders(t.Context(), expNumber, models.Order{
		UserID:     7,
		Number:     expNumber,
		Status:     models.PROCESSED,
		Accrual:    &expAccrual,
		UploadedAt: now,
	})
	assert.NoError(t, err)
	err = storage.Withdraw(t.Context(), 7, expNumber+"22", expBal)
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, err := storage.GetBalance(context.Background(), tt.id)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantBalance, balance)
		})
	}
}

func TestMemStorage_Withdraw(t *testing.T) {
	expAccrual := 300.5
	now := time.Now()
	tests := []struct {
		name    string
		id      int
		order   string
		bal     float64
		wantErr error
	}{
		{
			name:    "положительный тест",
			id:      7,
			order:   expNumber,
			bal:     expBal,
			wantErr: nil,
		},
		{
			name:    "недостаточно средств",
			id:      7,
			order:   expNumber + "22",
			bal:     expBal,
			wantErr: models.ErrInsufficientFunds,
		},
	}
	storage := memory.InitMemStorage()
	err := storage.SaveOrders(t.Context(), expNumber, models.Order{
		UserID:     7,
		Number:     expNumber,
		Status:     models.PROCESSED,
		Accrual:    &expAccrual,
		UploadedAt: now,
	})
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err = storage.Withdraw(t.Context(), tt.id, tt.order, tt.bal)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

func TestMemStorage_GetWithdraw(t *testing.T) {
	expAccrual := 500.5
	now := time.Now()
	tests := []struct {
		name    string
		id      int
		want    []models.Withdraw
		wantErr error
	}{
		{
			name: "положительный тест",
			id:   7,
			want: []models.Withdraw{
				{
					Order:        expNumber,
					Sum:          expBal,
					Processed_at: now,
				},
			},
			wantErr: nil,
		},
		{
			name:    "положительный тест",
			id:      5,
			want:    nil,
			wantErr: models.ErrNoWithdrawalsFound,
		},
	}
	storage := memory.InitMemStorage()
	err := storage.SaveOrders(t.Context(), expNumber, models.Order{
		UserID:     7,
		Number:     expNumber + "22",
		Status:     models.PROCESSED,
		Accrual:    &expAccrual,
		UploadedAt: now,
	})
	assert.NoError(t, err)
	err = storage.Withdraw(t.Context(), 7, expNumber, expBal)
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withdraw, err := storage.GetWithdraw(context.Background(), tt.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, withdraw, tt.want)

		})
	}
}
