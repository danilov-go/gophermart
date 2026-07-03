package db

import (
	"context"

	"github.com/danilov-go/gophermart/internal/models"
)

func (*storageDB) GetBalance(ctx context.Context, login string) (models.Balance, error) {
	return models.Balance{}, nil
}

func (*storageDB) Withdraw(ctx context.Context, login string, order string, bal float64) error {
	return nil
}

func (*storageDB) GetWithdraw(ctx context.Context, login string) ([]models.Withdraw, error) {
	return []models.Withdraw{}, nil
}
