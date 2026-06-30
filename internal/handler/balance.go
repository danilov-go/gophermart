package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/danilov-go/gophermart/internal/models"
)

type orderBalance struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func GetBalanceHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, login string, s storage) {
		balance, err := s.GetBalance(login)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(balance)
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

func WithdrawtBalanceHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, login string, s storage) {
		var buf bytes.Buffer
		var order orderBalance
		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = json.Unmarshal(buf.Bytes(), &order)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if order.Order == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = validLuna(order.Order)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		balance, err := s.GetBalance(login)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if balance.Current < order.Sum {
			w.WriteHeader(http.StatusPaymentRequired)
			return
		}
		err = s.Withdraw(login, order.Order, order.Sum)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func GetWithdrawalsBalanceHandler() LoginHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, login string, s storage) {
		withdraws, err := s.GetWithdraw(login)
		if err != nil {
			if errors.Is(err, models.ErrNoWithdrawalsFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		var buf bytes.Buffer
		err = json.NewEncoder(&buf).Encode(withdraws)
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
