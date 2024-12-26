package repositories

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

// SQLExecutor интерфейс с нужными функциями из sqlx.DB
type OldSQLExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (*sqlx.Tx, error)
}

// IRow Интерфейс для скрытия реализации sql.Row
type IRow interface {
	Scan(dest ...any) error
	Err() error
	StructScan(dest any) error
}

// ITX Интерфейс для скрытия реализации sql.TX
type ITX interface {
	PrepareNamed(query string) (INamedStmt, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	Rollback() error
	Commit() error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type IStmt interface {
	QueryRowxContext(ctx context.Context, arg interface{}) IRow
}

type INamedStmt interface {
	QueryRowxContext(ctx context.Context, arg interface{}) IRow
}

// SQLExecutor интерфейс с нужными функциями из sqlx.DB
type SQLExecutor interface {
	QueryRowContext(ctx context.Context, query string, args ...any) IRow
	PrepareNamed(query string) (INamedStmt, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) IRow
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	BeginTxx(ctx context.Context, opts *sql.TxOptions) (ITX, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}
