// Package handler реализует HTTP-интерфейс приложения.
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

// Handler связывает HTTP-запросов с хранилищем данных.
type Handler struct {
	storage Storage
	logger  log
}

// NewHandlers создает новый экземпляр Handler.
func NewHandlers(storage Storage, l log) *Handler {
	return &Handler{
		storage: storage,
		logger:  l,
	}
}

// Storage определяет методы для взаимодействия с хранилищем.
type Storage interface {
	Ping(ctx context.Context) error
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

// PGErrorClassification определяет категорию ошибки базы данных для повторных попыток выполнения.
type PGErrorClassification int

const (
	// NonRetriable определяет ошибку, которую нельзя исправить повторным запросом.
	NonRetriable PGErrorClassification = iota
	// Retriable определяет временную ошибку подключения, которую можно повторить.
	Retriable
)

// PostgresErrorClassifier проверяет типы ошибок PostgreSQL на возможность повтора операции.
type PostgresErrorClassifier struct{}

// NewPostgresErrorClassifier создает новый экземпляр классификатора ошибок.
func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

// Classify классифицирует ошибку для определения возможности повторных попыток.
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

// ErrorStorageMiddleware реализует механизма повторных попыток при сетевых сбоях.
type ErrorStorageMiddleware struct {
	next       Storage
	classifier *PostgresErrorClassifier
	duration   time.Duration
	interval   time.Duration
}

// NewErrorMiddleware создает новый экземпляр ErrorStorageMiddleware.
func NewErrorMiddleware(next Storage, duration, interval time.Duration) *ErrorStorageMiddleware {
	return &ErrorStorageMiddleware{
		next:       next,
		classifier: NewPostgresErrorClassifier(),
		duration:   duration,
		interval:   interval,
	}
}

func (rm *ErrorStorageMiddleware) replay(ctx context.Context, operation func() error) error {
	const maxRetries = 3
	duration := rm.duration
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}
		if attempt == maxRetries {
			return fmt.Errorf("попытки подключения исчерпаны: %w", err)
		}
		if rm.classifier.Classify(err) == Retriable {
			timer := time.NewTimer(duration)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
			timer.Stop()
			duration += rm.interval
			continue
		}
		return err
	}
	return nil
}

// SaveUser сохраняет нового пользователя в базе данных и возвращает его ID с механизмом повторных попыток.
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

// GetUser возвращает данные пользователя из базы данных с механизмом повторных попыток.
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

// SaveOrders сохраняет новый заказ в базе данных с механизмом повторных попыток.
func (rm *ErrorStorageMiddleware) SaveOrders(ctx context.Context, number string, orders models.Order) error {
	operation := func() error {
		return rm.next.SaveOrders(ctx, number, orders)
	}
	return rm.replay(ctx, operation)
}

// GetOrders возвращает отсортированный по времени список заказов пользователя с механизмом повторных попыток.
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

// GetBalance возвращает баланс и общую сумму списаний пользователя из базы данных с механизмом повторных попыток.
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

// Withdraw списывает баллы пользователя на указанный заказ в базе данных с механизмом повторных попыток.
func (rm *ErrorStorageMiddleware) Withdraw(ctx context.Context, id int, order string, bal float64) error {
	operation := func() error {
		return rm.next.Withdraw(ctx, id, order, bal)
	}
	return rm.replay(ctx, operation)
}

// GetWithdraw возвращает отсортированную по времени историю списаний пользователя из базы данных с механизмом повторных попыток.
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

// GetUnOrders возвращает список всех необработанных заказов с механизмом повторных попыток.
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

// UpdateStatus обновляет статус и сумму начисления для заказа с механизмом повторных попыток.
func (rm *ErrorStorageMiddleware) UpdateStatus(ctx context.Context, accrual models.Accrual) error {
	operation := func() error {
		return rm.next.UpdateStatus(ctx, accrual)
	}
	return rm.replay(ctx, operation)
}

// Ping выполняет проверку связи с базой данных.
func (rm *ErrorStorageMiddleware) Ping(ctx context.Context) error {
	return rm.next.Ping(ctx)
}
