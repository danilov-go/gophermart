package repository

import (
	"errors"
	"slices"

	"github.com/danilov-go/gophermart/internal/models"
)

type MemStorage struct {
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
	user, ok := m.users[login]
	if !ok {
		return models.User{}, errors.New("пользователь не найден")
	}
	return user, nil
}

func (m *MemStorage) SaveOrders(number string, orders models.Order) error {
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
