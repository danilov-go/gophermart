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

func (d *storageDB) SaveUser(ctx context.Context, login, passwordHash string) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var id int
	query := `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`
	err := d.db.QueryRowContext(ctx, query, login, passwordHash).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, errors.New("логин занят")
		}
		return 0, err
	}
	return id, nil
}

func (d *storageDB) GetUser(ctx context.Context, login string) (models.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var user models.User
	query := `SELECT id, login, password_hash FROM users WHERE login = $1`
	err := d.db.QueryRowContext(ctx, query, login).Scan(&user.ID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, errors.New("пользователь не найден")
		}
		return models.User{}, err
	}
	return user, nil
}
