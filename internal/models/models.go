// Package models определяет структуры данных и ошибки приложения.
package models

import (
	"errors"
	"time"
)

const (
	// NEW определяет статус, когда заказ загружен в систему, но не попал в обработку.
	NEW = "NEW"
	// PROCESSING определяет статус, когда вознаграждение за заказ рассчитывается.
	PROCESSING = "PROCESSING"
	// INVALID определяет статус, когда система расчёта вознаграждений отказала в расчёте.
	INVALID = "INVALID"
	// PROCESSED определяет статус, когда данные по заказу проверены и информация о расчёте успешно получена.
	PROCESSED = "PROCESSED"
)

var (
	// ErrOrderAlreadyUploadedBySameUser определяет ошибку, когда номер заказа уже был загружен этим пользователем.
	ErrOrderAlreadyUploadedBySameUser = errors.New("номер заказа уже был загружен этим пользователем")
	// ErrOrderAlreadyUploadedByOtherUser определяет ошибку, когда номер заказа уже был загружен другим пользователем.
	ErrOrderAlreadyUploadedByOtherUser = errors.New("номер заказа уже был загружен другим пользователем")
	// ErrNoOrdersFound определяет ошибку, когда у пользователя нет заказов.
	ErrNoOrdersFound = errors.New("у пользователя нет заказов")
	// ErrNoWithdrawalsFound определяет ошибку, когда нет ни одного списания.
	ErrNoWithdrawalsFound = errors.New("нет ни одного списания")
	// ErrUserAlreadyExists определяет ошибку, когда логин занят.
	ErrUserAlreadyExists = errors.New("логин занят")
	// ErrUserNotFound определяет ошибку, когда пользователь не найден.
	ErrUserNotFound = errors.New("пользователь не найден")
	// ErrUserNotFound определяет ошибку, когда недостаточно средств на счете.
	ErrInsufficientFunds = errors.New("недостаточно средств на счете")
)

// Order содержит данные о пользователе.
type User struct {
	// ID содержит уникальный идентификатор пользователя.
	ID int
	// Login содержит логин пользователя.
	Login string
	// PasswordHash содержит хешированный пароль пользователя.
	PasswordHash string
}

// Order содержит данные о заказе пользователя.
type Order struct {
	// UserID содержит уникальный идентификатор владельца заказа.
	UserID int `json:"-"`
	// Number содержит уникальный номер заказа.
	Number string `json:"number"`
	// Status содержит текущий статус обработки заказа.
	Status string `json:"status"`
	// Accrual содержит количество начисленных баллов за заказ.
	Accrual *float64 `json:"accrual,omitempty"`
	// UploadedAt содержит дату и время загрузки заказа.
	UploadedAt time.Time `json:"uploaded_at"`
}

// Orders определяет формат списка заказов для ответа.
type Orders struct {
	// Number содержит уникальный номер заказа.
	Number string `json:"number"`
	// Status содержит текущий статус обработки заказа.
	Status string `json:"status"`
	// Accrual содержит количество начисленных баллов за заказ.
	Accrual *float64 `json:"accrual,omitempty"`
	// UploadedAt содержит дату и время загрузки заказа.
	UploadedAt time.Time `json:"uploaded_at"`
}

// Balance содержит данные о состоянии счета пользователя.
type Balance struct {
	// Current содержит доступное количество баллов лояльности.
	Current float64 `json:"current"`
	// Withdrawn содержит общую сумму списанных баллов.
	Withdrawn float64 `json:"withdrawn"`
}

// Withdraw содержит данные о списании баллов.
type Withdraw struct {
	// Order содержит номер заказа, на который списываются баллы.
	Order string `json:"order"`
	// Sum содержит сумму списания.
	Sum float64 `json:"sum"`
	// ProcessedAt содержит дату и время проведения списания.
	Processed_at time.Time `json:"processed_at"`
}

// Accrual содержит структуру ответа от внешней системы начислений.
type Accrual struct {
	// Order содержит номер проверяемого заказа.
	Order string `json:"order"`
	// Status содержит статус заказа во внешней системе.
	Status string `json:"status"`
	// Accrual содержит сумму начисленных баллов.
	Accrual *float64 `json:"accrual,omitempty"`
}
