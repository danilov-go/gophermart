package handler_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/danilov-go/gophermart/internal/handler"
	"github.com/danilov-go/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

type testStorage struct {
	handler.Storage
	count     int
	returnErr error
}

func (m *testStorage) SaveUser(ctx context.Context, login, passwordHash string) (int, error) {
	m.count++
	return 1, m.returnErr
}

func (m *testStorage) GetUser(ctx context.Context, login string) (models.User, error) {
	m.count++
	return models.User{}, m.returnErr
}

func (m *testStorage) SaveOrders(ctx context.Context, number string, orders models.Order) error {
	m.count++
	return m.returnErr
}

func (m *testStorage) GetOrders(ctx context.Context, id int) ([]models.Orders, error) {
	m.count++
	return []models.Orders{}, m.returnErr
}

func (m *testStorage) GetBalance(ctx context.Context, id int) (models.Balance, error) {
	m.count++
	return models.Balance{}, m.returnErr
}

func (m *testStorage) Withdraw(ctx context.Context, id int, order string, bal float64) error {
	m.count++
	return m.returnErr
}

func (m *testStorage) GetWithdraw(ctx context.Context, id int) ([]models.Withdraw, error) {
	m.count++
	return []models.Withdraw{}, m.returnErr
}

func (m *testStorage) GetUnOrders(ctx context.Context) ([]models.Orders, error) {
	m.count++
	return []models.Orders{}, m.returnErr
}

func (m *testStorage) UpdateStatus(ctx context.Context, accrual models.Accrual) error {
	m.count++
	return m.returnErr
}

func (m *testStorage) Ping(ctx context.Context) error {
	m.count++
	return m.returnErr
}

func TestNewErrorMiddleware(t *testing.T) {
	retriableErr := &pgconn.PgError{
		Code: pgerrcode.ConnectionFailure,
	}
	nonRetriableErr := &pgconn.PgError{
		Code: pgerrcode.UniqueViolation,
	}
	expErr := fmt.Errorf("попытки подключения исчерпаны: %w", retriableErr)
	tests := []struct {
		name      string
		dbError   error
		wantCount int
		wantErr   error
	}{
		{
			name:      "с повтором",
			dbError:   retriableErr,
			wantCount: 4,
			wantErr:   expErr,
		},
		{
			name:      "без повтора",
			dbError:   nonRetriableErr,
			wantCount: 1,
			wantErr:   nonRetriableErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &testStorage{returnErr: tt.dbError}
			middleware := handler.NewErrorMiddleware(h, 1*time.Millisecond, 0)
			check := func(err error, expCount int, expErr error) {
				t.Helper()
				assert.Equal(t, expCount, h.count)
				assert.Error(t, err)
				assert.Equal(t, expErr.Error(), err.Error())
				h.count = 0
			}
			ctx := context.Background()
			_, err := middleware.SaveUser(ctx, expLogin, expPassword)
			check(err, tt.wantCount, tt.wantErr)
			_, err = middleware.GetUser(ctx, expLogin)
			check(err, tt.wantCount, tt.wantErr)
			err = middleware.SaveOrders(ctx, expNumber, models.Order{})
			check(err, tt.wantCount, tt.wantErr)
			_, err = middleware.GetOrders(ctx, 7)
			check(err, tt.wantCount, tt.wantErr)
			_, err = middleware.GetBalance(ctx, 7)
			check(err, tt.wantCount, tt.wantErr)
			err = middleware.Withdraw(ctx, 7, expNumber, expWithdraw)
			check(err, tt.wantCount, tt.wantErr)
			_, err = middleware.GetWithdraw(ctx, 7)
			check(err, tt.wantCount, tt.wantErr)
			_, err = middleware.GetUnOrders(ctx)
			check(err, tt.wantCount, tt.wantErr)
			err = middleware.UpdateStatus(ctx, models.Accrual{})
			check(err, tt.wantCount, tt.wantErr)
			err = middleware.Ping(ctx)
			check(err, 1, tt.dbError)
		})
	}
}
