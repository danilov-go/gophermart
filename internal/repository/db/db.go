package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type storageDB struct {
	db *sql.DB
}

func NewStorageDB(sql *sql.DB) storageDB {
	return storageDB{
		db: sql,
	}
}

func InitDB(ps string) (*storageDB, error) {
	db, err := sql.Open("pgx", ps)
	if err != nil {
		return nil, err
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		db.Close()
		return nil, err
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres",
		driver,
	)
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		db.Close()
		return nil, err
	}
	return &storageDB{
		db: db,
	}, nil
}

func (d *storageDB) Ping(ctx context.Context) error {
	err := d.db.PingContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
