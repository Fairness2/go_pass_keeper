package repositories

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
)

// DBAdapter Структура которая скрывает реализацию подключения к бд за функциями, которые можно замокировать
type DBAdapter struct {
	real *sqlx.DB
}

// NewDBAdapter Создание адаптера для sql.DB
func NewDBAdapter(real *sqlx.DB) *DBAdapter {
	return &DBAdapter{real}
}

func (D *DBAdapter) QueryRowContext(ctx context.Context, query string, args ...any) IRow {
	return D.real.QueryRowxContext(ctx, query, args...)
}

func (D *DBAdapter) PrepareNamed(query string) (INamedStmt, error) {
	stmt, err := D.real.PrepareNamed(query)
	if err != nil {
		return nil, err
	}
	return &NamedStmtAdapter{real: stmt}, err
}

func (D *DBAdapter) QueryRowxContext(ctx context.Context, query string, args ...interface{}) IRow {
	return D.real.QueryRowxContext(ctx, query, args...)
}

func (D *DBAdapter) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return D.real.SelectContext(ctx, dest, query, args...)
}

func (D *DBAdapter) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return D.real.NamedExecContext(ctx, query, arg)
}

func (D *DBAdapter) BeginTxx(ctx context.Context, opts *sql.TxOptions) (ITX, error) {
	tx, err := D.real.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &TxxAdapter{real: tx}, err
}

func (D *DBAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return D.real.ExecContext(ctx, query, args...)
}

// NamedStmtAdapter — это оболочка sqlx.NamedStmt для реализации интерфейса INamedStmt для мокабельного использования.
type NamedStmtAdapter struct {
	real *sqlx.NamedStmt
}

func (n *NamedStmtAdapter) QueryRowxContext(ctx context.Context, arg interface{}) IRow {
	return n.real.QueryRowxContext(ctx, arg)
}

// TxxAdapter — это структура, обеспечивающая реализацию интерфейса ITX путем оболочки экземпляра транзакции sqlx.Tx.
type TxxAdapter struct {
	real *sqlx.Tx
}

func (t *TxxAdapter) PrepareNamed(query string) (INamedStmt, error) {
	stmt, err := t.real.PrepareNamed(query)
	if err != nil {
		return nil, err
	}
	return &NamedStmtAdapter{real: stmt}, err
}

func (t *TxxAdapter) NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error) {
	return t.real.NamedExecContext(ctx, query, arg)
}

func (t *TxxAdapter) Rollback() error {
	return t.real.Rollback()
}

func (t *TxxAdapter) Commit() error {
	return t.real.Commit()
}

func (t *TxxAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.real.ExecContext(ctx, query, args...)
}
