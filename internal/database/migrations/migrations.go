package migrations

import (
	"database/sql"
	"embed"
	"github.com/pressly/goose/v3"
)

//go:embed sql/*.sql
var embedMigrations embed.FS

// Migrate применяет миграцию базы данных с использованием встроенных файлов миграции и диалекта PostgreSQL.
// Принимает соединение с базой данных (*sql.DB) в качестве входных данных и возвращает ошибку, если миграция не удалась.
func Migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "sql")
}
