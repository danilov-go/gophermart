// Package accrual реализует агент для периодического опроса внешней системы расчета баллов.
package accrual

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/go-resty/resty/v2"
)

type log interface {
	Errorw(msg string, keysAndValues ...any)
}

// Storage определяет методы для взаимодействия с хранилищем.
type Storage interface {
	GetUnOrders(ctx context.Context) ([]models.Orders, error)
	UpdateStatus(ctx context.Context, accrual models.Accrual) error
}

// Agent опрашивает внешний сервис расчета баллов.
type Agent struct {
	client       *resty.Client
	interval     int
	retryDefault int
	s            Storage
	logger       log
}

// New создает новый экземпляр Agent.
func New(interval, retryDefault int, addres string, l log, s Storage) *Agent {
	client := resty.New()
	client.SetTimeout(time.Second * 5)
	client.SetBaseURL(addres)
	return &Agent{
		client:       client,
		interval:     interval,
		retryDefault: retryDefault,
		s:            s,
		logger:       l,
	}
}

// Worker запускает фоновый процесс опроса статусов необработанных заказов.
func (a *Agent) Worker(ctx context.Context) error {
	duration := time.Duration(a.interval) * time.Second
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			orders, err := a.s.GetUnOrders(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					a.logger.Errorw("ошибка в получении заказов", "error", err)
					return nil
				}
				return err
			}
			if len(orders) == 0 {
				continue
			}
		Loop:
			for _, order := range orders {
				if errContext := ctx.Err(); errContext != nil {
					if errors.Is(errContext, context.Canceled) {
						return nil
					}
					return errContext
				}
				var accrual models.Accrual
				resp, err := a.client.R().
					SetContext(ctx).
					SetPathParam("number", order.Number).
					SetResult(&accrual).
					Get("/api/orders/{number}")
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return nil
					}
					a.logger.Errorw("ошибка в формировании запроса", "error", err)
					continue
				}
				var retryTime int
				if resp.StatusCode() == http.StatusTooManyRequests {
					retryAfter := resp.Header().Get("Retry-After")
					if retryAfter == "" {
						retryTime = a.retryDefault
					} else {
						timeAfter, errPars := strconv.Atoi(retryAfter)
						if errPars != nil {
							a.logger.Errorw("ошибка в преобразовании retryTime в число", "error", errPars)
							retryTime = a.retryDefault
						} else {
							retryTime = timeAfter
						}
					}
					timer := time.NewTimer(time.Duration(retryTime) * time.Second)
					select {
					case <-timer.C:
					case <-ctx.Done():
						timer.Stop()
						return nil
					}
					timer.Stop()
					break Loop
				}
				if resp.StatusCode() != http.StatusOK {
					a.logger.Errorw("статус ответа от accrual", "status", resp.StatusCode())
					continue
				}
				err = a.s.UpdateStatus(ctx, accrual)
				if err != nil {
					a.logger.Errorw("ошибка обновления статуса", "error", err)
					break Loop
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}
