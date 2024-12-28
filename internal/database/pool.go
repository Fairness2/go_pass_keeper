package database

import (
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"passkeeper/internal/logger"
)

var (
	ErrorEmptyDSN                    = errors.New("empty dsn")
	ErrorIncorrectMaxConnections     = errors.New("incorrect max connections")
	ErrorIncorrectMaxIdleConnections = errors.New("incorrect max idle connections")
)

// DBPool глобальный пул подключений к базе данных для приложения c функцией закрытия
type DBPool struct {
	DBx *sqlx.DB
}

// newPgDBx инициализирует новый пул соединений PostgreSQL, используя заданный DSN, максимальное количество открытых соединений и максимальное количество простаивающих соединений.
// Он проверяет соединение с базой данных во время инициализации, выполняя операцию Ping.
// Возвращает объект *sqlx.DB в случае успеха или ошибки, если соединение или проверка не удались.
func newPgDBx(dsn string, maxConnections int, maxIdleConnections int) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(maxConnections)
	db.SetMaxIdleConns(maxIdleConnections)
	// Если дсн не передан, то просто возвращаем созданный пул, он не работоспособен

	// Сразу проверим работоспособность соединения
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// NewDB инициализация подключения к бд
func NewDB(dsn string, maxConnections int, maxIdleConnections int) (*DBPool, error) {
	if dsn == "" {
		return nil, ErrorEmptyDSN
	}
	if maxConnections <= 0 {
		return nil, ErrorIncorrectMaxConnections
	}
	if maxIdleConnections <= 0 {
		return nil, ErrorIncorrectMaxIdleConnections
	}
	// Создание пула подключений к базе данных для приложения
	db, err := newPgDBx(dsn, maxConnections, maxIdleConnections)
	if err != nil {
		return nil, err
	}
	pool := &DBPool{
		DBx: db,
	}

	return pool, nil
}

/*func (p *DBPool) Migrate() error {
	logger.Log.Info("Migrate migrations")
	// Применим миграции
	return migrations.Migrate(p.DBx.DB)
}*/

// Close закрытие базы данных
func (p *DBPool) Close() {
	logger.Log.Info("Closing database connection for defer")
	if p.DBx != nil {
		err := p.DBx.Close()
		if err != nil {
			logger.Log.Error(err)
		}
	}
}
