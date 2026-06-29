package models

import (
	"errors"
	"time"
)

const (
	NEW        = "NEW"
	PROCESSING = "PROCESSING"
	INVALID    = "INVALID"
	PROCESSED  = "PROCESSED"
)

type User struct {
	ID           int
	Login        string
	PasswordHash string
}
type Order struct {
	Login      string    `json:"login"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Orders struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    float64   `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

var (
	ErrOrderAlreadyUploadedBySameUser  = errors.New("номер заказа уже был загружен этим пользователем")
	ErrOrderAlreadyUploadedByOtherUser = errors.New("номер заказа уже был загружен другим пользователем")
	ErrNoOrdersFound                   = errors.New("у пользователя нет заказов")
)
