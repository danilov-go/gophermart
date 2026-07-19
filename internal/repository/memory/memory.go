// Package memory реализует хранилище данных в оперативной памяти.
package memory

import (
	"context"
	"sync"

	"github.com/danilov-go/gophermart/internal/models"
)

// MemStorage реализует хранилище данных в оперативной памяти.
type MemStorage struct {
	mu          sync.RWMutex
	nextID      int
	userId      map[int]models.User
	userLogin   map[string]int
	orders      map[string]models.Order
	withdrawals map[int][]models.Withdraw
}

// InitMemStorage создает новый экземпляр MemStorage.
func InitMemStorage() *MemStorage {
	return &MemStorage{
		nextID:      1,
		userId:      make(map[int]models.User),
		userLogin:   make(map[string]int),
		orders:      make(map[string]models.Order),
		withdrawals: make(map[int][]models.Withdraw),
	}
}

// Ping проверяет доступность хранилища.
func (m *MemStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}
