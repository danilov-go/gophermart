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

const expNumber = "12345678903"

func Test_storageDB_SaveOrders(t *testing.T) {
	queryInsert := `INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)`
	querySelect := `SELECT user_id FROM orders WHERE number = $1`
	now := time.Now().Truncate(time.Second)
	tests := []struct {
		name      string
		orderNum  string
		order     models.Order
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name:     "положительный тест",
			orderNum: expNumber,
			order: models.Order{
				UserID:     1,
				Status:     models.NEW,
				UploadedAt: now,
			},
			setupMock: func(mock sqlmock.Sqlmock) {

				mock.ExpectExec(queryInsert).
					WithArgs(1, expNumber, models.NEW, now).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: nil,
		},
		{
			name:     "номер заказа уже был загружен этим пользователем",
			orderNum: expNumber,
			order:    models.Order{UserID: 1, Status: models.NEW, UploadedAt: now},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(queryInsert).
					WithArgs(1, expNumber, models.NEW, now).
					WillReturnError(errPostgresUnique)
				mock.ExpectQuery(querySelect).
					WithArgs(expNumber).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(1))
			},
			wantErr: models.ErrOrderAlreadyUploadedBySameUser,
		},
		{
			name:     "номер заказа уже был загружен другим пользователем",
			orderNum: expNumber,
			order:    models.Order{UserID: 1, Status: models.NEW, UploadedAt: now},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(queryInsert).
					WithArgs(1, expNumber, models.NEW, now).
					WillReturnError(errPostgresUnique)
				mock.ExpectQuery(querySelect).
					WithArgs(expNumber).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(2))
			},
			wantErr: models.ErrOrderAlreadyUploadedByOtherUser,
		},
		{
			name:     "ошибка базы данных",
			orderNum: expNumber,
			order:    models.Order{UserID: 1, Status: models.NEW, UploadedAt: now},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(queryInsert).
					WithArgs(1, expNumber, models.NEW, now).
					WillReturnError(errTest)
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
			err = storage.SaveOrders(t.Context(), tt.orderNum, tt.order)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetOrders(t *testing.T) {
	query := `SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`
	now := time.Now().Truncate(time.Second)
	expAccrual := 450.50
	tests := []struct {
		name      string
		userID    int
		setupMock func(mock sqlmock.Sqlmock)
		want      []models.Orders
		wantErr   error
	}{
		{
			name:   "положительный тест",
			userID: 7,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
					AddRow(expNumber, models.PROCESSED, &expAccrual, now).
					AddRow(expNumber+"22", models.NEW, nil, now)
				mock.ExpectQuery(query).
					WithArgs(7).
					WillReturnRows(rows)
			},
			want: []models.Orders{
				{Number: expNumber, Status: models.PROCESSED, Accrual: &expAccrual, UploadedAt: now},
				{Number: expNumber + "22", Status: models.NEW, Accrual: nil, UploadedAt: now},
			},
			wantErr: nil,
		},
		{
			name:   "заказы не найдены",
			userID: 5,
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"})
				mock.ExpectQuery(query).
					WithArgs(5).
					WillReturnRows(rows)
			},
			want:    []models.Orders{},
			wantErr: models.ErrNoOrdersFound,
		},
		{
			name:   "ошибка базы данных",
			userID: 5,
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(5).
					WillReturnError(errTest)
			},
			want:    []models.Orders{},
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
			order, err := storage.GetOrders(t.Context(), tt.userID)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, order)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetUnOrders(t *testing.T) {
	query := `SELECT number, status, accrual, uploaded_at FROM orders WHERE status = $1 OR status = $2`
	now := time.Now().Truncate(time.Second)
	expAccrual := 200.5
	tests := []struct {
		name      string
		setupMock func(mock sqlmock.Sqlmock)
		want      []models.Orders
		wantErr   error
	}{
		{
			name: "положительный тест",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"}).
					AddRow(expNumber, models.NEW, nil, now).
					AddRow(expNumber+"22", models.PROCESSING, &expAccrual, now)
				mock.ExpectQuery(query).
					WithArgs(models.NEW, models.PROCESSING).
					WillReturnRows(rows)
			},
			want: []models.Orders{
				{Number: expNumber, Status: models.NEW, Accrual: nil, UploadedAt: now},
				{Number: expNumber + "22", Status: models.PROCESSING, Accrual: &expAccrual, UploadedAt: now},
			},
			wantErr: nil,
		},
		{
			name: "заказов нет",
			setupMock: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"number", "status", "accrual", "uploaded_at"})
				mock.ExpectQuery(query).
					WithArgs(models.NEW, models.PROCESSING).
					WillReturnRows(rows)
			},
			want:    []models.Orders{},
			wantErr: nil,
		},
		{
			name: "ошибка базы данных",
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(query).
					WithArgs(models.NEW, models.PROCESSING).
					WillReturnError(errTest)
			},
			want:    []models.Orders{},
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
			order, err := storage.GetUnOrders(t.Context())
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.want, order)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_UpdateStatus(t *testing.T) {
	queryOrders := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 RETURNING user_id`
	queryUsers := `UPDATE users SET current = current + $1 WHERE id = $2`
	expAccrual := 200.5
	tests := []struct {
		name      string
		accrual   models.Accrual
		setupMock func(mock sqlmock.Sqlmock)
		wantErr   error
	}{
		{
			name:    "положительный тест",
			accrual: models.Accrual{Order: expNumber, Status: models.PROCESSED, Accrual: &expAccrual},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryOrders).
					WithArgs(models.PROCESSED, &expAccrual, expNumber).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
				mock.ExpectExec(queryUsers).
					WithArgs(expAccrual, 7).
					WillReturnResult(sqlmock.NewResult(0, 1))
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name:    "положительный тест",
			accrual: models.Accrual{Order: expNumber, Status: models.PROCESSING, Accrual: nil},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryOrders).
					WithArgs(models.PROCESSING, nil, expNumber).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
				mock.ExpectCommit()
			},
			wantErr: nil,
		},
		{
			name:    "заказ не найден",
			accrual: models.Accrual{Order: expNumber + "22", Status: models.PROCESSED, Accrual: &expAccrual},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryOrders).
					WithArgs(models.PROCESSED, &expAccrual, expNumber+"22").
					WillReturnError(sql.ErrNoRows)
				mock.ExpectRollback()
			},
			wantErr: models.ErrNoOrdersFound,
		},
		{
			name: "ошибка базы данных",
			accrual: models.Accrual{
				Order:   expNumber,
				Status:  models.PROCESSED,
				Accrual: &expAccrual,
			},
			setupMock: func(mock sqlmock.Sqlmock) {
				mock.ExpectBegin()
				mock.ExpectQuery(queryOrders).
					WithArgs(models.PROCESSED, &expAccrual, expNumber).
					WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(7))
				mock.ExpectExec(queryUsers).
					WithArgs(expAccrual, 7).
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
			err = storage.UpdateStatus(t.Context(), tt.accrual)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
