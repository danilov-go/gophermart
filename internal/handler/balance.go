package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/danilov-go/gophermart/internal/models"
)

type orderBalance struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func (h *BalanceHandler) GetBalanceHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, user AuthUser) {
		ctx := r.Context()
		balance, err := h.storage.GetBalance(ctx, user.ID)
		if err != nil {
			if errors.Is(err, models.ErrUserNotFound) {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			h.logger.Errorw("ошибка получения баланса", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(balance)
		if err != nil {
			h.logger.Errorw("ошибка сериализации", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf.Bytes())
		if err != nil {
			h.logger.Errorw("ошибка отправки данных", "error", err)
			return
		}
	}
}

func (h *BalanceHandler) WithdrawtBalanceHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, user AuthUser) {
		ctx := r.Context()
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
			h.logger.Errorw("не соответствие content-type", "content-type", r.Header.Get("Content-Type"))
			http.Error(w, http.StatusText(http.StatusUnsupportedMediaType), http.StatusUnsupportedMediaType)
			return
		}
		var buf bytes.Buffer
		var order orderBalance
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			h.logger.Errorw("ошибка чтения тела запроса", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(buf.Bytes(), &order)
		if err != nil {
			h.logger.Errorw("ошибка десилиризации", "error", err)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		if order.Order == "" {
			h.logger.Errorw("пустой номер заказа", "error", errors.New("получен пустой номер заказа"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = validLuna(order.Order)
		if err != nil {
			h.logger.Errorw("ошибка валидации номера заказа", "error", err)
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		err = h.storage.Withdraw(ctx, user.ID, order.Order, order.Sum)
		if err != nil {
			if errors.Is(err, models.ErrInsufficientFunds) {
				h.logger.Errorw("ошибка списания баланса", "error", err)
				http.Error(w, http.StatusText(http.StatusPaymentRequired), http.StatusPaymentRequired)
				return
			}
			h.logger.Errorw("ошибка списания", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *BalanceHandler) GetWithdrawalsBalanceHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, user AuthUser) {
		ctx := r.Context()
		withdraws, err := h.storage.GetWithdraw(ctx, user.ID)
		if err != nil {
			if errors.Is(err, models.ErrNoWithdrawalsFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h.logger.Errorw("ошибка списания", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(withdraws)
		if err != nil {
			h.logger.Errorw("ошибка сериализации", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf.Bytes())
		if err != nil {
			h.logger.Errorw("ошибка отправки данных", "error", err)
			return
		}
	}
}
