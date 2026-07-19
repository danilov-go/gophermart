package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/stretchr/testify/assert"
)

const secretKey = "secret_key"

func TestAuthMiddleware(t *testing.T) {
	token, err := handler.BuildJWTString(7, expLogin, secretKey)
	assert.NoError(t, err)

	tests := []struct {
		name       string
		key        string
		authHeader string
		code       int
		wantUser   handler.AuthUser
		wantErr    bool
	}{
		{
			name:       "положительный тест",
			key:        secretKey,
			authHeader: "Bearer " + token,
			code:       http.StatusOK,
			wantUser:   handler.AuthUser{ID: 7, Login: expLogin},
			wantErr:    false,
		},
		{
			name:       "пустой заголовок",
			key:        secretKey,
			authHeader: "",
			code:       http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "невалидный токен",
			key:        secretKey,
			authHeader: "Bearer broken-token-data",
			code:       http.StatusUnauthorized,
			wantErr:    true,
		},
		{
			name:       "неверный ключ",
			key:        "unknow",
			authHeader: "Bearer " + token,
			code:       http.StatusUnauthorized,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var expUser handler.AuthUser
			next := func(w http.ResponseWriter, r *http.Request, user handler.AuthUser) {
				expUser = user
				w.WriteHeader(http.StatusOK)
			}
			h := handler.AuthMiddleware(tt.key, next)
			req := httptest.NewRequest(http.MethodGet, "/test-auth", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			res := httptest.NewRecorder()
			h.ServeHTTP(res, req)
			assert.Equal(t, tt.code, res.Code)

			if !tt.wantErr {
				assert.Equal(t, tt.wantUser, expUser)
			}
		})
	}
}
