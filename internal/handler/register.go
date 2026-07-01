package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

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

func (h *BalanceHandler) RegisterUser(key string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := decode(r)
		if err != nil {
			h.logger.Errorw("ошибка десилиризации", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if user.Login == "" || user.Password == "" {
			h.logger.Errorw("пустой логин или пароль", "login", user.Login, "password", user.Password)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		_, err = h.storage.GetUser(user.Login)
		if err == nil {
			h.logger.Errorw("пользователь с таким именем уже существует", "error", err)
			http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
			return
		}
		hashStringPassword := hash(user.Password, key)
		_, err = h.storage.SaveUser(user.Login, hashStringPassword)
		if err != nil {
			h.logger.Errorw("ошибка сохранения пользователя", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		signedToken, err := buildJWTString(user.Login, key)
		if err != nil {
			h.logger.Errorw("ошибка аутентификации", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", "Bearer "+signedToken)
		w.WriteHeader(http.StatusOK)
	})
}

func (h *BalanceHandler) LoginUser(key string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, err := decode(r)
		if err != nil {
			h.logger.Errorw("ошибка десериализации", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if user.Login == "" || user.Password == "" {
			h.logger.Errorw("пустой логин или пароль", "login", user.Login)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		store, err := h.storage.GetUser(user.Login)
		if err != nil {
			h.logger.Errorw("неверный логин или пароль", "error", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		hashStringPassword := hash(user.Password, key)
		if !hmac.Equal([]byte(hashStringPassword), []byte(store.PasswordHash)) {
			h.logger.Errorw("пароли не совпадают")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized) // 401
			return
		}
		signedToken, err := buildJWTString(user.Login, key)
		if err != nil {
			h.logger.Errorw("ошибка аутентификации", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", "Bearer "+signedToken)
		w.WriteHeader(http.StatusOK)
	})
}
