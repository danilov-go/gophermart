package memory

import (
	"context"
	"errors"

	"github.com/danilov-go/gophermart/internal/models"
)

func (m *MemStorage) SaveUser(ctx context.Context, login, passwordHash string) (int, error) {
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

func (m *MemStorage) GetUser(ctx context.Context, login string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, ok := m.users[login]
	if !ok {
		return models.User{}, errors.New("пользователь не найден")
	}
	return user, nil
}
