package memory

import (
	"context"
	"slices"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
)

func (m *MemStorage) GetBalance(ctx context.Context, id int) (models.Balance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var userBalance models.Balance
	var countW float64
	var countA float64
	for _, order := range m.orders {
		if order.UserID == id && order.Status == models.PROCESSED {
			if order.Accrual != nil {
				countA += *order.Accrual
			}
		}
	}
	for _, w := range m.withdrawals[id] {
		countW += w.Sum
	}
	userBalance = models.Balance{
		Current:   countA - countW,
		Withdrawn: countW,
	}
	return userBalance, nil
}

func (m *MemStorage) Withdraw(ctx context.Context, id int, order string, bal float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	var countW float64
	var countA float64
	for _, order := range m.orders {
		if order.UserID == id && order.Status == models.PROCESSED {
			if order.Accrual != nil {
				countA += *order.Accrual
			}
		}
	}
	for _, w := range m.withdrawals[id] {
		countW += w.Sum
	}
	balance := countA - countW
	if balance < bal {
		return models.ErrInsufficientFunds
	}
	withdraw := models.Withdraw{
		Order:        order,
		Sum:          bal,
		Processed_at: time.Now(),
	}
	m.withdrawals[id] = append(m.withdrawals[id], withdraw)
	return nil
}

func (m *MemStorage) GetWithdraw(ctx context.Context, id int) ([]models.Withdraw, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	withdraw, ok := m.withdrawals[id]
	if !ok {
		return nil, models.ErrNoWithdrawalsFound
	}
	slices.SortFunc(withdraw, func(i, j models.Withdraw) int {
		return i.Processed_at.Compare(j.Processed_at)
	})
	return withdraw, nil
}
