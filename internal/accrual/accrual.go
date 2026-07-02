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

type storage interface {
	GetUnOrders() ([]models.Orders, error)
	UpdateStatus(accrual models.Accrual) error
}

type Agent struct {
	Client   *resty.Client
	interval int
	s        storage
	logger   log
}

func New(interval int, addres string, l log, s storage) *Agent {
	client := resty.New()
	client.SetTimeout(time.Second * 5)
	client.SetBaseURL(addres)
	return &Agent{
		Client:   client,
		interval: interval,
		s:        s,
		logger:   l,
	}
}

func (a *Agent) Worker(ctx context.Context) {
	duration := time.Duration(a.interval) * time.Second
	ticker := time.NewTicker(duration)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			orders, err := a.s.GetUnOrders()
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
				resp, err := a.Client.R().
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
				err = a.s.UpdateStatus(accrual)
				if err != nil {
					a.logger.Errorw("ошибка обновления статуса", "error", err)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
