package memory

import (
	"sync"

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
