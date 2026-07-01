package handler

import "github.com/danilov-go/gophermart/internal/models"

type storage interface {
	SaveUser(login, passwordHash string) (int, error)
	GetUser(login string) (models.User, error)
	SaveOrders(number string, orders models.Order) error
	GetOrders(login string) ([]models.Orders, error)
	GetBalance(login string) (models.Balance, error)
	Withdraw(login string, order string, bal float64) error
	GetWithdraw(login string) ([]models.Withdraw, error)
}
type log interface {
	Errorw(msg string, keysAndValues ...any)
}

type BalanceHandler struct {
	storage storage
	logger  log
}

func NewHandlers(storage storage, l log) *BalanceHandler {
	return &BalanceHandler{
		storage: storage,
		logger:  l,
	}
}
