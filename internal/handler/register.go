package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Login string `json:"login"`
	jwt.RegisteredClaims
}

const tokenExp = time.Hour * 24

type storage interface {
	SaveUser(login, passwordHash string) (int, error)
	GetUser(login string) (models.User, error)
	SaveOrders(number string, orders models.Order) error
	GetOrders(login string) ([]models.Orders, error)
	GetBalance(login string) (models.Balance, error)
}

type loginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func buildJWTString(login, key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenExp)),
		},
		Login: login,
	})
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func hash(password, key string) string {
	hs := hmac.New(sha256.New, []byte(key))
	hs.Write([]byte(password))
	hashedPassword := hs.Sum(nil)
	return hex.EncodeToString(hashedPassword[:])
}

func decode(r *http.Request) (loginPassword, error) {
	var buf bytes.Buffer
	var user loginPassword
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		return loginPassword{}, err
	}
	err = json.Unmarshal(buf.Bytes(), &user)
	if err != nil {
		return loginPassword{}, err
	}
	return user, nil
}

func RegisterUser(s storage, key string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := decode(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if user.Login == "" || user.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = s.GetUser(user.Login)
		if err == nil {
			w.WriteHeader(http.StatusConflict)
			return
		}
		hashStringPassword := hash(user.Password, key)
		_, err = s.SaveUser(user.Login, hashStringPassword)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		signedToken, err := buildJWTString(user.Login, key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", "Bearer "+signedToken)
		w.WriteHeader(http.StatusOK)
	})
}

func LoginUser(s storage, key string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := decode(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if user.Login == "" || user.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		store, err := s.GetUser(user.Login)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		hashStringPassword := hash(user.Password, key)
		if !hmac.Equal([]byte(hashStringPassword), []byte(store.PasswordHash)) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		signedToken, err := buildJWTString(user.Login, key)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", "Bearer "+signedToken)
		w.WriteHeader(http.StatusOK)
	})
}
