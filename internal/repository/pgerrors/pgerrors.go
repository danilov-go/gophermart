package pgerrors

import (
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type PGErrorClassification int

const (
	NonRetriable PGErrorClassification = iota
	Retriable
)

type PostgresErrorClassifier struct{}

func NewPostgresErrorClassifier() *PostgresErrorClassifier {
	return &PostgresErrorClassifier{}
}

func (c *PostgresErrorClassifier) Classify(err error) PGErrorClassification {
	if err == nil {
		return NonRetriable
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return СlassifyPgError(pgErr)
	}
	return NonRetriable
}

func СlassifyPgError(pgErr *pgconn.PgError) PGErrorClassification {
	switch pgErr.Code {
	case pgerrcode.ConnectionException,
		pgerrcode.ConnectionDoesNotExist,
		pgerrcode.ConnectionFailure:
		return Retriable
	case pgerrcode.TransactionRollback,
		pgerrcode.SerializationFailure,
		pgerrcode.DeadlockDetected:
		return Retriable
	case pgerrcode.CannotConnectNow:
		return Retriable
	}
	switch pgErr.Code {
	case pgerrcode.DataException,
		pgerrcode.NullValueNotAllowedDataException:
		return NonRetriable
	case pgerrcode.IntegrityConstraintViolation,
		pgerrcode.RestrictViolation,
		pgerrcode.NotNullViolation,
		pgerrcode.ForeignKeyViolation,
		pgerrcode.UniqueViolation,
		pgerrcode.CheckViolation:
		return NonRetriable
	case pgerrcode.SyntaxErrorOrAccessRuleViolation,
		pgerrcode.SyntaxError,
		pgerrcode.UndefinedColumn,
		pgerrcode.UndefinedTable,
		pgerrcode.UndefinedFunction:
		return NonRetriable
	}
	return NonRetriable
}
