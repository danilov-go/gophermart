package db_test

import (
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/danilov-go/gophermart/internal/repository/db"
	"github.com/stretchr/testify/assert"
)

var errTest = errors.New("error test")

func Test_storageDB_Ping(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
	}{
		{
			name:    "положительный тест",
			wantErr: nil,
		},
		{
			name:    "ошибка базы данных",
			wantErr: errTest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
			assert.NoError(t, err)
			defer sql.Close()
			exp := mock.ExpectPing()
			if tt.wantErr != nil {
				exp.WillReturnError(tt.wantErr)
			}
			storage := db.NewStorageDB(sql)
			err = storage.Ping(t.Context())
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}
