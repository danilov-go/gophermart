package memory

import (
	"sync"

	"github.com/danilov-go/gophermart/internal/models"
)

type MemStorage struct {
	mu          sync.RWMutex
	nextID      int
	userId      map[int]models.User
	userLogin   map[string]int
	orders      map[string]models.Order
	withdrawals map[int][]models.Withdraw
}

func InitMemStorage() *MemStorage {
	return &MemStorage{
		nextID:      1,
		userId:      make(map[int]models.User),
		userLogin:   make(map[string]int),
		orders:      make(map[string]models.Order),
		withdrawals: make(map[int][]models.Withdraw),
	}
}
