package db

import (
	"context"

	"github.com/danilov-go/gophermart/internal/models"
)

func (*storageDB) SaveOrders(ctx context.Context, number string, orders models.Order) error {
	return nil
}

func (*storageDB) GetOrders(ctx context.Context, login string) ([]models.Orders, error) {
	return []models.Orders{}, nil
}

func (*storageDB) GetUnOrders(ctx context.Context) ([]models.Orders, error) {
	return []models.Orders{}, nil
}

func (*storageDB) UpdateStatus(ctx context.Context, accrual models.Accrual) error {
	return nil
}
