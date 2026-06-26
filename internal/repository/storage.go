package repository

import (
	"errors"

	"github.com/danilov-go/gophermart/internal/models"
)

type MemStorage struct {
	users  map[string]models.User
	nextID int
}

func InitMemStorage() *MemStorage {
	return &MemStorage{
		users:  make(map[string]models.User),
		nextID: 1,
	}
}

func (m *MemStorage) SaveUser(login, passwordHash string) (int, error) {
	if _, ok := m.users[login]; ok {
		return 0, errors.New("пользователь с таким именем уже существует")
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
		return models.User{}, errors.New("пользователь с таким именем отсутствует")
	}
	return user, nil
}
