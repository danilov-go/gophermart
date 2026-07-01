package repository

import (
	"errors"
	"slices"
	"sync"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
)

type MemStorage struct {
	mu          sync.RWMutex
	nextID      int
	users       map[string]models.User
	orders      map[string]models.Order
	withdrawals map[string][]models.Withdraw
}

func InitMemStorage() *MemStorage {
	return &MemStorage{
		nextID:      1,
		users:       make(map[string]models.User),
		orders:      make(map[string]models.Order),
		withdrawals: make(map[string][]models.Withdraw),
	}
}

func (m *MemStorage) SaveUser(login, passwordHash string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[login]; ok {
		return 0, errors.New("логин занят")
	}
	userID := m.nextID
	newUser := models.User{
		ID:           userID,
		Login:        login,
		PasswordHash: passwordHash,
	}
	m.users[login] = newUser
	m.nextID++
	return userID, nil
}

func (m *MemStorage) GetUser(login string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, ok := m.users[login]
	if !ok {
		return models.User{}, errors.New("пользователь не найден")
	}
	return user, nil
}

func (m *MemStorage) SaveOrders(number string, orders models.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if order, ok := m.orders[number]; ok {
		if order.Login == orders.Login {
			return models.ErrOrderAlreadyUploadedBySameUser
		}
		return models.ErrOrderAlreadyUploadedByOtherUser

	}
	m.orders[number] = orders
	return nil
}

func (m *MemStorage) GetOrders(login string) ([]models.Orders, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var userOrders []models.Orders
	for number, order := range m.orders {
		if order.Login == login {
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

func (m *MemStorage) GetBalance(login string) (models.Balance, error) {
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

func (m *MemStorage) Withdraw(login string, order string, bal float64) error {
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

func (m *MemStorage) GetWithdraw(login string) ([]models.Withdraw, error) {
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

func (m *MemStorage) GetUnOrders() ([]models.Orders, error) {
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

func (m *MemStorage) UpdateStatus(accrual models.Accrual) error {
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
