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

// QueryRowContext выполняет запрос, который, как ожидается, вернет одну строку, и сопоставляет его с IRow для дальнейшей обработки.
func (D *DBAdapter) QueryRowContext(ctx context.Context, query string, args ...any) IRow {
	return D.real.QueryRowxContext(ctx, query, args...)
}

// PrepareNamed подготавливает именованный запрос к выполнению и возвращает INamedStmt или ошибку, если подготовка не удалась.
func (D *DBAdapter) PrepareNamed(query string) (INamedStmt, error) {
	stmt, err := D.real.PrepareNamed(query)
	if err != nil {
		return nil, err
	}
	return &NamedStmtAdapter{real: stmt}, err
}

// QueryRowxContext выполняет запрос, который, как ожидается, вернет одну строку, и сопоставляет его с IRow для дальнейшей обработки.
func (D *DBAdapter) QueryRowxContext(ctx context.Context, query string, args ...any) IRow {
	return D.real.QueryRowxContext(ctx, query, args...)
}

// SelectContext выполняет запрос и сканирует полученные строки в целевой объект, используя заданный контекст.
func (D *DBAdapter) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	return D.real.SelectContext(ctx, dest, query, args...)
}

// NamedExecContext выполняет именованный оператор с предоставленным контекстом и аргументом, возвращая результат или ошибку.
func (D *DBAdapter) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return D.real.NamedExecContext(ctx, query, arg)
}

// BeginTxx начинает новую транзакцию, используя предоставленный контекст и параметры транзакции, возвращая ITX или ошибку.
func (D *DBAdapter) BeginTxx(ctx context.Context, opts *sql.TxOptions) (ITX, error) {
	tx, err := D.real.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &TxxAdapter{real: tx}, err
}

// ExecContext выполняет запрос, не возвращая никаких строк, используя данный контекст для управления отменой или тайм-аутом.
func (D *DBAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return D.real.ExecContext(ctx, query, args...)
}

// NamedStmtAdapter — это оболочка sqlx.NamedStmt для реализации интерфейса INamedStmt для мокабельного использования.
type NamedStmtAdapter struct {
	real *sqlx.NamedStmt
}

// QueryRowxContext выполняет именованный оператор с контекстом и аргументами, возвращая одну строку в качестве интерфейса IRow.
func (n *NamedStmtAdapter) QueryRowxContext(ctx context.Context, arg any) IRow {
	return n.real.QueryRowxContext(ctx, arg)
}

// TxxAdapter — это структура, обеспечивающая реализацию интерфейса ITX путем оболочки экземпляра транзакции sqlx.Tx.
type TxxAdapter struct {
	real *sqlx.Tx
}

// PrepareNamed подготавливает именованный оператор для выполнения и возвращает интерфейс INamedStmt или ошибку, если подготовка не удалась.
func (t *TxxAdapter) PrepareNamed(query string) (INamedStmt, error) {
	stmt, err := t.real.PrepareNamed(query)
	if err != nil {
		return nil, err
	}
	return &NamedStmtAdapter{real: stmt}, err
}

// NamedExecContext выполняет именованный оператор с заданным контекстом и аргументом, возвращая sql.Result или ошибку.
func (t *TxxAdapter) NamedExecContext(ctx context.Context, query string, arg any) (sql.Result, error) {
	return t.real.NamedExecContext(ctx, query, arg)
}

// Rollback откатывает транзакцию, отменяя все изменения, внесенные во время транзакции.
func (t *TxxAdapter) Rollback() error {
	return t.real.Rollback()
}

// Commit фиксирует текущую транзакцию, применяя все изменения, внесенные за время существования транзакции.
func (t *TxxAdapter) Commit() error {
	return t.real.Commit()
}

// ExecContext выполняет запрос с предоставленным контекстом и аргументами, возвращая результат или ошибку в случае неудачи.
func (t *TxxAdapter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.real.ExecContext(ctx, query, args...)
}
