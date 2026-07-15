package db_test

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danilov-go/gophermart/internal/models"
	"github.com/danilov-go/gophermart/internal/repository/db"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

const expLogin = "login"
const expPassword = "password"

var errPostgresUnique = &pgconn.PgError{
	Code: pgerrcode.UniqueViolation,
}

func Test_storageDB_SaveUser(t *testing.T) {
	tests := []struct {
		name         string
		login        string
		passwordHash string
		mockRows     *sqlmock.Rows
		mockErr      error
		wantID       int
		wantErr      error
	}{
		{
			name:         "положительный тест",
			login:        expLogin,
			passwordHash: expPassword,
			mockRows:     sqlmock.NewRows([]string{"id"}).AddRow(7),
			mockErr:      nil,
			wantID:       7,
			wantErr:      nil,
		},
		{
			name:         "логин занят",
			login:        expLogin + "1",
			passwordHash: expPassword,
			mockRows:     nil,
			mockErr:      errPostgresUnique,
			wantID:       0,
			wantErr:      models.ErrUserAlreadyExists,
		},
		{
			name:         "ошибка базы данных",
			login:        expLogin + "2",
			passwordHash: expPassword,
			mockRows:     nil,
			mockErr:      errTest,
			wantID:       0,
			wantErr:      errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer sql.Close()
			query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`
			exp := mock.ExpectQuery(query).WithArgs(tt.login, tt.passwordHash)
			if tt.mockErr != nil {
				exp.WillReturnError(tt.mockErr)
			}
			if tt.mockRows != nil {
				exp.WillReturnRows(tt.mockRows)
			}
			storage := db.NewStorageDB(sql)
			id, err := storage.SaveUser(t.Context(), tt.login, tt.passwordHash)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantID, id)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func Test_storageDB_GetUser(t *testing.T) {
	tests := []struct {
		name     string
		login    string
		mockRows *sqlmock.Rows
		mockErr  error
		wantUser models.User
		wantErr  error
	}{
		{
			name:  "положительный тест",
			login: expLogin,
			mockRows: sqlmock.NewRows([]string{"id", "login", "password_hash"}).
				AddRow(1, expLogin, expPassword),
			mockErr: nil,
			wantUser: models.User{
				ID:           1,
				Login:        expLogin,
				PasswordHash: expPassword,
			},
			wantErr: nil,
		},
		{
			name:     "пользователь не найден",
			login:    expLogin,
			mockRows: nil,
			mockErr:  sql.ErrNoRows,
			wantUser: models.User{},
			wantErr:  models.ErrUserNotFound,
		},
		{
			name:     "ошибка базы данных",
			login:    expLogin,
			mockRows: nil,
			mockErr:  errTest,
			wantUser: models.User{},
			wantErr:  errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			assert.NoError(t, err)
			defer sql.Close()
			query := `SELECT id, login, password_hash FROM users WHERE login = $1`
			exp := mock.ExpectQuery(query).WithArgs(tt.login)
			if tt.mockErr != nil {
				exp.WillReturnError(tt.mockErr)
			}
			if tt.mockRows != nil {
				exp.WillReturnRows(tt.mockRows)
			}
			storage := db.NewStorageDB(sql)
			user, err := storage.GetUser(t.Context(), tt.login)
			assert.ErrorIs(t, err, tt.wantErr)
			assert.Equal(t, tt.wantUser, user)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
