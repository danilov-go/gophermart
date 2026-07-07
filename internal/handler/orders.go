package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
)

func validLuna(number string) error {
	sum := 0
	lengthEven := len(number) % 2
	for i, num := range number {
		n, err := strconv.Atoi(string(num))
		if err != nil {
			return errors.New("номер заказа содержит некорректные символы")
		}
		if i%2 == lengthEven {
			n *= 2
			if n > 9 {
				n -= 9
			}
		}
		sum += n
	}
	if sum%10 != 0 {
		return errors.New("номер заказа не прошёл проверку по алгоритму Луна")
	}
	return nil
}

func (h *BalanceHandler) SaveOrderHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, user AuthUser) {
		ctx := r.Context()
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
			h.logger.Errorw("не соответствие content-type", "content-type", r.Header.Get("Content-Type"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			h.logger.Errorw("ошибка чтения тела запроса", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		number := strings.TrimSpace(string(body))
		if number == "" {
			h.logger.Errorw("пустой номер заказа", "error", errors.New("получен пустой номер заказа"))
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}
		err = validLuna(number)
		if err != nil {
			h.logger.Errorw("ошибка валидации номера заказа", "error", err)
			http.Error(w, http.StatusText(http.StatusUnprocessableEntity), http.StatusUnprocessableEntity)
			return
		}
		orderSave := models.Order{
			UserID:     user.ID,
			Status:     models.NEW,
			UploadedAt: time.Now(),
		}
		err = h.storage.SaveOrders(ctx, number, orderSave)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrOrderAlreadyUploadedBySameUser):
				w.WriteHeader(http.StatusOK)
				return
			case errors.Is(err, models.ErrOrderAlreadyUploadedByOtherUser):
				http.Error(w, http.StatusText(http.StatusConflict), http.StatusConflict)
				return
			default:
				h.logger.Errorw("ошибка сохранения заказа", "error", err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func (h *BalanceHandler) GetOrderHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, user AuthUser) {
		ctx := r.Context()
		orders, err := h.storage.GetOrders(ctx, user.ID)
		if err != nil {
			if errors.Is(err, models.ErrNoOrdersFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			h.logger.Errorw("ошибка получения заказа", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(orders)
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
