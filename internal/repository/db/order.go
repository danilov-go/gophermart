package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// SaveOrders сохраняет новый заказ в базе данных.
func (d *storageDB) SaveOrders(ctx context.Context, number string, orders models.Order) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `INSERT INTO orders (user_id, number, status, uploaded_at) VALUES ($1, $2, $3, $4)`
	_, err := d.db.ExecContext(ctx, query, orders.UserID, number, orders.Status, orders.UploadedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			var id int
			query = `SELECT user_id FROM orders WHERE number = $1`
			errCheck := d.db.QueryRowContext(ctx, query, number).Scan(&id)
			if errCheck != nil {
				return errCheck
			}
			if id == orders.UserID {
				return models.ErrOrderAlreadyUploadedBySameUser
			}
			return models.ErrOrderAlreadyUploadedByOtherUser
		}
		return err
	}
	return nil
}

// GetOrders возвращает отсортированный по времени список заказов пользователя.
func (d *storageDB) GetOrders(ctx context.Context, id int) ([]models.Orders, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	orders := make([]models.Orders, 0)
	query := `SELECT number, status, accrual, uploaded_at FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`
	rows, err := d.db.QueryContext(ctx, query, id)
	if err != nil {
		return orders, err
	}
	defer rows.Close()
	for rows.Next() {
		var order models.Orders
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return orders, models.ErrNoOrdersFound
	}
	return orders, nil
}

// GetUnOrders возвращает список всех необработанных заказов.
func (d *storageDB) GetUnOrders(ctx context.Context) ([]models.Orders, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	orders := make([]models.Orders, 0)
	query := `SELECT number, status, accrual, uploaded_at FROM orders WHERE status = $1 OR status = $2`
	rows, err := d.db.QueryContext(ctx, query, models.NEW, models.PROCESSING)
	if err != nil {
		return orders, err
	}
	defer rows.Close()
	for rows.Next() {
		var order models.Orders
		err = rows.Scan(&order.Number, &order.Status, &order.Accrual, &order.UploadedAt)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

// UpdateStatus обновляет статус и сумму начисления для заказа.
func (d *storageDB) UpdateStatus(ctx context.Context, accrual models.Accrual) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID int
	queryOrder := `UPDATE orders SET status = $1, accrual = $2 WHERE number = $3 RETURNING user_id`
	err = tx.QueryRowContext(ctx, queryOrder, accrual.Status, accrual.Accrual, accrual.Order).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrNoOrdersFound
		}
		return err
	}
	if accrual.Status == models.PROCESSED && accrual.Accrual != nil {
		queryUser := `UPDATE users SET current = current + $1 WHERE id = $2`
		_, err = tx.ExecContext(ctx, queryUser, *accrual.Accrual, userID)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
