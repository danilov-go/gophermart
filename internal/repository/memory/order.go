package memory

import (
	"context"
	"slices"

	"github.com/danilov-go/gophermart/internal/models"
)

func (m *MemStorage) SaveOrders(ctx context.Context, number string, newOrder models.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if order, ok := m.orders[number]; ok {
		if order.UserID == newOrder.UserID {
			return models.ErrOrderAlreadyUploadedBySameUser
		}
		return models.ErrOrderAlreadyUploadedByOtherUser

	}
	m.orders[number] = newOrder
	return nil
}

func (m *MemStorage) GetOrders(ctx context.Context, id int) ([]models.Orders, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var userOrders []models.Orders
	for number, order := range m.orders {
		if order.UserID == id {
			user := models.Orders{
				Number:     number,
				Status:     order.Status,
				Accrual:    order.Accrual,
				UploadedAt: order.UploadedAt,
			}
			userOrders = append(userOrders, user)
		}
	}
	if len(userOrders) == 0 {
		return userOrders, models.ErrNoOrdersFound
	}
	slices.SortFunc(userOrders, func(i, j models.Orders) int {
		return i.UploadedAt.Compare(j.UploadedAt)
	})
	return userOrders, nil
}

func (m *MemStorage) GetUnOrders(ctx context.Context) ([]models.Orders, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var orders []models.Orders
	for number, order := range m.orders {
		if order.Status == models.NEW || order.Status == models.PROCESSING {
			user := models.Orders{
				Number:     number,
				Status:     order.Status,
				Accrual:    order.Accrual,
				UploadedAt: order.UploadedAt,
			}
			orders = append(orders, user)
		}
	}
	return orders, nil
}

func (m *MemStorage) UpdateStatus(ctx context.Context, accrual models.Accrual) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	order, ok := m.orders[accrual.Order]
	if !ok {
		return models.ErrNoOrdersFound
	}
	order.Status = accrual.Status
	order.Accrual = accrual.Accrual
	m.orders[accrual.Order] = order
	return nil
}
