package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/danilov-go/gophermart/internal/repository/memory"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

type want struct {
	code int
}

type loginPassword struct {
	Login    string
	Password string
}

var key = "secret"
var expLogin = "login1"
var expPassword = "123"

func TestRegisterUser(t *testing.T) {
	tests := []struct {
		name string
		user loginPassword
		want want
	}{
		{
			name: "положительный тест",
			user: loginPassword{
				Login:    expLogin,
				Password: expPassword,
			},
			want: want{
				code: 200,
			},
		},
		{
			name: "пустой пароль",
			user: loginPassword{
				Login:    "login2",
				Password: "",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "пустой логин",
			user: loginPassword{
				Login:    "",
				Password: "123",
			},
			want: want{
				code: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			r := chi.NewRouter()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			r.Post("/register", h.RegisterUser(key))
			body, err := json.Marshal(tt.user)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, tt.want.code, rec.Code)
			if rec.Code == http.StatusOK {
				user, err := storage.GetUser(t.Context(), tt.user.Login)
				require.NoError(t, err)
				password := handler.Hash(tt.user.Password, key)
				assert.Equal(t, password, user.PasswordHash)
				assert.Equal(t, 1, user.ID)
				assert.Equal(t, tt.user.Login, user.Login)
				authHeader := rec.Header().Get("Authorization")
				require.NotEmpty(t, authHeader)
				assert.Contains(t, authHeader, "Bearer ")
				tokenString := bytes.TrimPrefix([]byte(authHeader), []byte("Bearer "))
				expUser, err := handler.GetUserLogin(string(tokenString), key)
				require.NoError(t, err)
				assert.Equal(t, tt.user.Login, expUser.Login)
				assert.Equal(t, 1, expUser.ID)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	tests := []struct {
		name string
		user loginPassword
		want want
	}{
		{
			name: "положительный тест",

			user: loginPassword{
				Login:    expLogin,
				Password: expPassword,
			},
			want: want{
				code: 200,
			},
		},
		{
			name: "неверный пароль",

			user: loginPassword{
				Login:    expLogin,
				Password: "321",
			},
			want: want{
				code: 401,
			},
		},
		{
			name: "пустой пароль",

			user: loginPassword{
				Login:    expLogin,
				Password: "",
			},
			want: want{
				code: 400,
			},
		},
		{
			name: "неверный логин",
			user: loginPassword{
				Login:    "login2",
				Password: expPassword,
			},
			want: want{
				code: 401,
			},
		},
		{
			name: "пустой логин",
			user: loginPassword{
				Login:    "",
				Password: expPassword,
			},
			want: want{
				code: 400,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := memory.InitMemStorage()
			password := handler.Hash(expPassword, key)
			_, err := storage.SaveUser(t.Context(), expLogin, password)
			require.NoError(t, err)
			r := chi.NewRouter()
			logger := zaptest.NewLogger(t)
			h := handler.NewHandlers(storage, logger.Sugar())
			r.Post("/login", h.LoginUser(key))
			body, err := json.Marshal(tt.user)
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.Equal(t, tt.want.code, rec.Code)
			if rec.Code == http.StatusOK {
				user, err := storage.GetUser(t.Context(), tt.user.Login)
				require.NoError(t, err)
				password := handler.Hash(tt.user.Password, key)
				assert.Equal(t, password, user.PasswordHash)
				assert.Equal(t, tt.user.Login, user.Login)
				authHeader := rec.Header().Get("Authorization")
				require.NotEmpty(t, authHeader)
				assert.Contains(t, authHeader, "Bearer ")
				tokenString := bytes.TrimPrefix([]byte(authHeader), []byte("Bearer "))
				expUser, err := handler.GetUserLogin(string(tokenString), key)
				require.NoError(t, err)
				assert.Equal(t, tt.user.Login, expUser.Login)
				assert.Equal(t, 1, expUser.ID)
			}
		})
	}
}
