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
	UserID     int       `json:"-"`
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Orders struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdraw struct {
	Order        string    `json:"order"`
	Sum          float64   `json:"sum"`
	Processed_at time.Time `json:"processed_at"`
}
type Accrual struct {
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}

var (
	ErrOrderAlreadyUploadedBySameUser  = errors.New("номер заказа уже был загружен этим пользователем")
	ErrOrderAlreadyUploadedByOtherUser = errors.New("номер заказа уже был загружен другим пользователем")
	ErrNoOrdersFound                   = errors.New("у пользователя нет заказов")
	ErrNoWithdrawalsFound              = errors.New("нет ни одного списания")
	ErrUserAlreadyExists               = errors.New("логин занят")
	ErrUserNotFound                    = errors.New("пользователь не найден")
	ErrInsufficientFunds               = errors.New("недостаточно средств на счете")
)
