package handler

import (
	"context"

	"github.com/danilov-go/gophermart/internal/models"
)

type Storage interface {
	SaveUser(ctx context.Context, login, passwordHash string) (int, error)
	GetUser(ctx context.Context, login string) (models.User, error)
	SaveOrders(ctx context.Context, number string, orders models.Order) error
	GetOrders(ctx context.Context, login string) ([]models.Orders, error)
	GetBalance(ctx context.Context, login string) (models.Balance, error)
	Withdraw(ctx context.Context, login string, order string, bal float64) error
	GetWithdraw(ctx context.Context, login string) ([]models.Withdraw, error)
	GetUnOrders(ctx context.Context) ([]models.Orders, error)
	UpdateStatus(ctx context.Context, accrual models.Accrual) error
}

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type BalanceHandler struct {
	storage Storage
	logger  log
}

func NewHandlers(storage Storage, l log) *BalanceHandler {
	return &BalanceHandler{
		storage: storage,
		logger:  l,
	}
}
