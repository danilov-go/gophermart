package memory

import (
	"context"

	"github.com/danilov-go/gophermart/internal/models"
)

// SaveUser сохраняет нового пользователя в оперативной памяти и возвращает его ID.
func (m *MemStorage) SaveUser(ctx context.Context, login, passwordHash string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.userLogin[login]; ok {
		return 0, models.ErrUserAlreadyExists
	}
	userID := m.nextID
	m.nextID++
	newUser := models.User{
		ID:           userID,
		Login:        login,
		PasswordHash: passwordHash,
	}
	m.userId[userID] = newUser
	m.userLogin[login] = userID
	return userID, nil
}

// GetUser возвращает данные пользователя из оперативной памяти.
func (m *MemStorage) GetUser(ctx context.Context, login string) (models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	userID, ok := m.userLogin[login]
	if !ok {
		return models.User{}, models.ErrUserNotFound
	}
	user, ok := m.userId[userID]
	if !ok {
		return models.User{}, models.ErrUserNotFound
	}
	return user, nil
}
