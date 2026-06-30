package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

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
