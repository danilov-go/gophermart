package memory

import (
	"context"
	"slices"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
)

func (m *MemStorage) GetBalance(ctx context.Context, login string) (models.Balance, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var userBalance models.Balance
	var countW float64
	var countA float64
	for _, order := range m.orders {
		if order.Login == login && order.Status == models.PROCESSED {
			countA += order.Accrual
		}
	}
	for _, w := range m.withdrawals[login] {
		countW += w.Sum
	}
	userBalance = models.Balance{
		Current:   countA - countW,
		Withdrawn: countW,
	}
	return userBalance, nil
}

func (m *MemStorage) Withdraw(ctx context.Context, login string, order string, bal float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	withdraw := models.Withdraw{
		Order:        order,
		Sum:          bal,
		Processed_at: time.Now(),
	}
	m.withdrawals[login] = append(m.withdrawals[login], withdraw)
	return nil
}

func (m *MemStorage) GetWithdraw(ctx context.Context, login string) ([]models.Withdraw, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	withdraw, ok := m.withdrawals[login]
	if !ok {
		return []models.Withdraw{}, models.ErrNoWithdrawalsFound
	}
	slices.SortFunc(withdraw, func(i, j models.Withdraw) int {
		return i.Processed_at.Compare(j.Processed_at)
	})
	return withdraw, nil
}
