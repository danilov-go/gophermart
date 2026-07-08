package handler

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

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

type Storage interface {
	SaveUser(ctx context.Context, login, passwordHash string) (int, error)
	GetUser(ctx context.Context, login string) (models.User, error)
	SaveOrders(ctx context.Context, number string, orders models.Order) error
	GetOrders(ctx context.Context, id int) ([]models.Orders, error)
	GetBalance(ctx context.Context, id int) (models.Balance, error)
	Withdraw(ctx context.Context, id int, order string, bal float64) error
	GetWithdraw(ctx context.Context, id int) ([]models.Withdraw, error)
	GetUnOrders(ctx context.Context) ([]models.Orders, error)
	UpdateStatus(ctx context.Context, accrual models.Accrual) error
}

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return classifyPgError(pgErr)
	}
	return NonRetriable
}

func classifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable
	}
	return NonRetriable
}

type ErrorStorageMiddleware struct {
	next       Storage
	classifier *PostgresErrorClassifier
}

func NewErrorMiddleware(next Storage) *ErrorStorageMiddleware {
	return &ErrorStorageMiddleware{
		next:       next,
		classifier: NewPostgresErrorClassifier(),
	}
}

func (rm *ErrorStorageMiddleware) replay(ctx context.Context, operation func() error) error {
	const maxRetries = 3
	duration := 1
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		if attempt == maxRetries {
			return fmt.Errorf("попытки подключения исчерпаны: %w", err)
		}
		if rm.classifier.Classify(err) == Retriable {
			timer := time.NewTimer(time.Duration(duration) * time.Second)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			timer.Stop()
			duration += 2
			continue
		}
		return err
	}
	return nil
}

func (rm *ErrorStorageMiddleware) SaveUser(ctx context.Context, login, passwordHash string) (int, error) {
	var id int
	operation := func() error {
		var err error
		id, err = rm.next.SaveUser(ctx, login, passwordHash)
		return err
	}
	err := rm.replay(ctx, operation)
	return id, err
}

func (rm *ErrorStorageMiddleware) GetUser(ctx context.Context, login string) (models.User, error) {
	var user models.User
	operation := func() error {
		var err error
		user, err = rm.next.GetUser(ctx, login)
		return err
	}
	err := rm.replay(ctx, operation)
	return user, err
}

func (rm *ErrorStorageMiddleware) SaveOrders(ctx context.Context, number string, orders models.Order) error {
	operation := func() error {
		return rm.next.SaveOrders(ctx, number, orders)
	}
	return rm.replay(ctx, operation)
}

func (rm *ErrorStorageMiddleware) GetOrders(ctx context.Context, id int) ([]models.Orders, error) {
	var orders []models.Orders
	operation := func() error {
		var err error
		orders, err = rm.next.GetOrders(ctx, id)
		return err
	}
	err := rm.replay(ctx, operation)
	return orders, err
}

func (rm *ErrorStorageMiddleware) GetBalance(ctx context.Context, id int) (models.Balance, error) {
	var balance models.Balance
	operation := func() error {
		var err error
		balance, err = rm.next.GetBalance(ctx, id)
		return err
	}
	err := rm.replay(ctx, operation)
	return balance, err
}

func (rm *ErrorStorageMiddleware) Withdraw(ctx context.Context, id int, order string, bal float64) error {
	operation := func() error {
		return rm.next.Withdraw(ctx, id, order, bal)
	}
	return rm.replay(ctx, operation)
}

func (rm *ErrorStorageMiddleware) GetWithdraw(ctx context.Context, id int) ([]models.Withdraw, error) {
	var withdraw []models.Withdraw
	operation := func() error {
		var err error
		withdraw, err = rm.next.GetWithdraw(ctx, id)
		return err
	}
	err := rm.replay(ctx, operation)
	return withdraw, err
}

func (rm *ErrorStorageMiddleware) GetUnOrders(ctx context.Context) ([]models.Orders, error) {
	var orders []models.Orders
	operation := func() error {
		var err error
		orders, err = rm.next.GetUnOrders(ctx)
		return err
	}
	err := rm.replay(ctx, operation)
	return orders, err
}

func (rm *ErrorStorageMiddleware) UpdateStatus(ctx context.Context, accrual models.Accrual) error {
	operation := func() error {
		return rm.next.UpdateStatus(ctx, accrual)
	}
	return rm.replay(ctx, operation)
}
