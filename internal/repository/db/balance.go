package db

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/danilov-go/gophermart/internal/models"
)

// GetBalance возвращает баланс и общую сумму списаний пользователя из базы данных.
func (d *storageDB) GetBalance(ctx context.Context, id int) (models.Balance, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var user models.Balance
	query := `SELECT current, withdrawn FROM users WHERE id = $1`
	err := d.db.QueryRowContext(ctx, query, id).Scan(&user.Current, &user.Withdrawn)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Balance{}, models.ErrUserNotFound
		}
		return models.Balance{}, err
	}
	return user, nil
}

// Withdraw списывает баллы пользователя на указанный заказ в базе данных.
func (d *storageDB) Withdraw(ctx context.Context, id int, order string, bal float64) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var userID int
	query := `UPDATE users SET current = current - $2, withdrawn = withdrawn + $2 WHERE id = $1 AND current >= $2 RETURNING id`
	err = tx.QueryRowContext(ctx, query, id, bal).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrInsufficientFunds
		}
		return err
	}
	query = `INSERT INTO withdrawals (user_id, order_number, sum) VALUES ($1, $2, $3)`
	_, err = tx.ExecContext(ctx, query, userID, order, bal)
	if err != nil {
		return err
	}
	return tx.Commit()
}

// GetWithdraw возвращает отсортированную по времени историю списаний пользователя из базы данных.
func (d *storageDB) GetWithdraw(ctx context.Context, id int) ([]models.Withdraw, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	query := `SELECT order_number, sum, processed_at FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`
	rows, err := d.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	withdrawals := make([]models.Withdraw, 0)
	for rows.Next() {
		var withdrawal models.Withdraw
		err = rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, withdrawal)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return withdrawals, nil
}
