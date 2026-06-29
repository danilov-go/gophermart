package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
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

func SaveOrderHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, login string, s storage) {
		if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		number := strings.TrimSpace(string(body))
		if number == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = validLuna(number)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		orderSave := models.Order{
			Login:      login,
			Status:     models.NEW,
			UploadedAt: time.Now(),
		}
		err = s.SaveOrders(number, orderSave)
		if err != nil {
			switch {
			case errors.Is(err, models.ErrOrderAlreadyUploadedBySameUser):
				w.WriteHeader(http.StatusOK)
				return
			case errors.Is(err, models.ErrOrderAlreadyUploadedByOtherUser):
				w.WriteHeader(http.StatusConflict)
				return
			default:
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func GetOrderHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, login string, s storage) {
		orders, err := s.GetOrders(login)
		if err != nil {
			if errors.Is(err, models.ErrNoOrdersFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(orders)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(buf.Bytes())
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
