// Package accrual реализует агент для периодического опроса внешней системы расчета баллов.
package accrual

import (
	"context"
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
	client   *resty.Client
	interval int
	s        Storage
	logger   log
}

// New создает новый экземпляр Agent.
func New(interval int, addres string, l log, s Storage) *Agent {
	client := resty.New()
	client.SetTimeout(time.Second * 5)
	client.SetBaseURL(addres)
	return &Agent{
		client:   client,
		interval: interval,
		s:        s,
		logger:   l,
	}
}

// Worker запускает фоновый процесс опроса статусов необработанных заказов.
func (a *Agent) Worker(ctx context.Context) {
	duration := time.Duration(a.interval) * time.Second
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			orders, err := a.s.GetUnOrders(ctx)
			if err != nil {
				a.logger.Errorw("ошибка в получении заказов", "error", err)
				continue
			}
			if len(orders) == 0 {
				continue
			}
			for _, order := range orders {
				if ctx.Err() != nil {
					return
				}
				var accrual models.Accrual
				resp, err := a.client.R().
					SetContext(ctx).
					SetPathParam("number", order.Number).
					SetResult(&accrual).
					Get("/api/orders/{number}")
				if err != nil {
					a.logger.Errorw("ошибка в формировании запроса", "error", err)
					continue
				}
				if resp.StatusCode() == http.StatusTooManyRequests {
					retryAfter := resp.Header().Get("Retry-After")
					if retryAfter == "" {
						break
					}
					retryTime, err := strconv.Atoi(retryAfter)
					if err != nil {
						a.logger.Errorw("ошибка в преобразовании retryTime в число", "error", err)
					}
					select {
					case <-time.After(time.Duration(retryTime) * time.Second):
					case <-ctx.Done():
						return
					}
					break
				}
				if resp.StatusCode() != http.StatusOK {
					continue
				}
				err = a.s.UpdateStatus(ctx, accrual)
				if err != nil {
					a.logger.Errorw("ошибка обновления статуса", "error", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
