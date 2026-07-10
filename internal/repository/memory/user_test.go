package memory_test

import (
	"testing"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/danilov-go/gophermart/internal/repository/memory"
	"github.com/stretchr/testify/assert"
)

const expLogin = "login"
const expPassword = "password"

func TestMemStorage_SaveUser(t *testing.T) {
	tests := []struct {
		name         string
		login        string
		passwordHash string
		id           int
		wantErr      bool
	}{
		{
			name:         "положительный тест",
			login:        expLogin,
			passwordHash: expPassword,
			id:           1,
			wantErr:      false,
		},
		{
			name:         "логин занят",
			login:        expLogin,
			passwordHash: expPassword,
			id:           0,
			wantErr:      true,
		},
	}
	storage := memory.InitMemStorage()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := storage.SaveUser(t.Context(), tt.login, tt.passwordHash)
			assert.Equal(t, tt.id, id)
			if tt.wantErr {
				assert.ErrorIs(t, err, models.ErrUserAlreadyExists)
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestMemStorage_GetUser(t *testing.T) {
	tests := []struct {
		name    string
		login   string
		want    models.User
		wantErr bool
	}{
		{
			name:  "положительный тест",
			login: expLogin,
			want: models.User{
				ID:           1,
				Login:        expLogin,
				PasswordHash: expPassword,
			},
			wantErr: false,
		},
		{
			name:    "пользователь не найден",
			login:   "unknown",
			want:    models.User{},
			wantErr: true,
		},
	}
	storage := memory.InitMemStorage()
	_, err := storage.SaveUser(t.Context(), expLogin, expPassword)
	assert.NoError(t, err)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := storage.GetUser(t.Context(), tt.login)
			if tt.wantErr {
				assert.ErrorIs(t, err, models.ErrUserNotFound)
				assert.Equal(t, tt.want, user)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, user)
			}
		})
	}
}
