package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	ID    int    `json:"id"`
	Login string `json:"login"`
	jwt.RegisteredClaims
}

type AuthUser struct {
	ID    int
	Login string
}

const tokenExp = time.Hour * 24

type LoginHandlerFunc func(w http.ResponseWriter, r *http.Request, user AuthUser)

func GetUserLogin(tokenString, key string) (AuthUser, error) {
	var user AuthUser
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(key), nil
	})
	if err != nil {
		return AuthUser{}, err
	}
	if !token.Valid {
		return AuthUser{}, errors.New("token is not valid")
	}
	user = AuthUser{
		ID:    claims.ID,
		Login: claims.Login,
	}
	return user, nil
}

func AuthMiddleware(key string, h LoginHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(token, "Bearer ")
		user, err := GetUserLogin(tokenString, key)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		h(w, r, user)
	}
}
