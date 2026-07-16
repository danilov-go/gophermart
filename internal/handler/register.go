package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type loginPassword struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// BuildJWTString создает строку JWT-токена для указанного пользователя.
func BuildJWTString(id int, login, key string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenExp)),
		},
		ID:    id,
		Login: login,
	})
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// Hash возвращает HMAC-SHA256 хеш пароля.
func Hash(password, key string) string {
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

// RegisterUser возвращает обработчик для регистрации нового пользователя.
func (h *Handler) RegisterUser(key string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
		hashStringPassword := Hash(user.Password, key)
		id, err := h.storage.SaveUser(ctx, user.Login, hashStringPassword)
		if err != nil {
			if errors.Is(err, models.ErrUserAlreadyExists) {
				h.logger.Errorw("пользователь с таким именем уже существует", "error", err)
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			}
			h.logger.Errorw("ошибка сохранения пользователя", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		signedToken, err := BuildJWTString(id, user.Login, key)
		if err != nil {
			h.logger.Errorw("ошибка аутентификации", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", "Bearer "+signedToken)
		w.WriteHeader(http.StatusOK)
	})
}

// LoginUser возвращает обработчик для аутентификации существующего пользователя.
func (h *Handler) LoginUser(key string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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
		store, err := h.storage.GetUser(ctx, user.Login)
		if err != nil {
			h.logger.Errorw("неверный логин или пароль", "error", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		hashStringPassword := Hash(user.Password, key)
		if !hmac.Equal([]byte(hashStringPassword), []byte(store.PasswordHash)) {
			h.logger.Errorw("пароли не совпадают")
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized) // 401
			return
		}
		signedToken, err := BuildJWTString(store.ID, user.Login, key)
		if err != nil {
			h.logger.Errorw("ошибка аутентификации", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Authorization", "Bearer "+signedToken)
		w.WriteHeader(http.StatusOK)
	})
}
