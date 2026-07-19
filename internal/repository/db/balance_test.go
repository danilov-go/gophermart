package db_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danilov-go/gophermart/internal/models"
	"github.com/danilov-go/gophermart/internal/repository/db"
	"github.com/stretchr/testify/assert"
)

func Test_storageDB_GetBalance(t *testing.T) {
	query := `SELECT current, withdrawn FROM users WHERE id = $1`
	tests := []struct {
		name        string
		id          int
		setupMock   func(mock sqlmock.Sqlmock)
		wantBalance models.Balance
		wantErr     error
	}{
		{
			name: "положительный тест",
			id:   7,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"current", "withdrawn"}).AddRow(500.5, 200.5)
				mock.ExpectQuery(query).
					WithArgs(7).
					WillReturnRows(rows)
			},
			wantBalance: models.Balance{
				Current:   500.5,
				Withdrawn: 200.5,
			},
			wantErr: nil,
		},
		{
			name: "пользователь не найден",
			id:   5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(5).
					WillReturnError(sql.ErrNoRows)
			},
			wantBalance: models.Balance{},
			wantErr:     models.ErrUserNotFound,
		},
		{
			name: "ошибка базы данных",
			id:   5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(5).
					WillReturnError(errTest)
			},
			wantBalance: models.Balance{},
			wantErr:     errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer sql.Close()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			balance, err := storage.GetBalance(t.Context(), tt.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantBalance, balance)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_Withdraw(t *testing.T) {
	queryUpdate := `UPDATE users SET current = current - $2, withdrawn = withdrawn + $2 WHERE id = $1 AND current >= $2 RETURNING id`
	queryInsert := `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`
	expBal := 200.5
	tests := []struct {
		name      string
		id        int
		order     string
		bal       float64
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name:  "положительный тест",
			id:    7,
			order: expNumber,
			bal:   expBal,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryUpdate).
					WithArgs(7, expBal).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
				mock.ExpectExec(queryInsert).
					WithArgs(7, expNumber, expBal).
					WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name:  "недостаточно средств",
			id:    7,
			order: expNumber,
			bal:   500.1,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryUpdate).
					WithArgs(7, 500.1).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectRollback()
			},
			wantErr: models.ErrInsufficientFunds,
		},
		{
			name:  "ошибка базы данных",
			id:    7,
			order: expNumber,
			bal:   expBal,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryUpdate).
					WithArgs(7, expBal).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))
				mock.ExpectExec(queryInsert).
					WithArgs(7, expNumber, expBal).
					WillReturnError(errTest)

				mock.ExpectRollback()
			},
			wantErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer sql.Close()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			err = storage.Withdraw(t.Context(), tt.id, tt.order, tt.bal)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetWithdraw(t *testing.T) {
	query := `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name      string
		id        int
		setupMock func(mock sqlmock.Sqlmock)
		want      []models.Withdraw
		wantErr   error
	}{
		{
			name: "положительный тест",
			id:   7,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"}).
					AddRow(expNumber, 100.0, now).
					AddRow(expNumber+"22", 200.5, now)
				mock.ExpectQuery(query).
					WithArgs(7).
					WillReturnRows(rows)
			},
			want: []models.Withdraw{
				{Order: expNumber, Sum: 100.0, ProcessedAt: now},
				{Order: expNumber + "22", Sum: 200.5, ProcessedAt: now},
			},
			wantErr: nil,
		},
		{
			name: "положительный тест",
			id:   5,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"order_number", "sum", "processed_at"})
				mock.ExpectQuery(query).
					WithArgs(5).
					WillReturnRows(rows)
			},
			want:    []models.Withdraw{},
			wantErr: nil,
		},
		{
			name: "ошибка базы данных",
			id:   5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(5).
					WillReturnError(errTest)
			},
			want:    nil,
			wantErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer sql.Close()
			tt.setupMock(mock)
			storage := db.NewStorageDB(sql)
			got, err := storage.GetWithdraw(t.Context(), tt.id)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, got)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
