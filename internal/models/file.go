package models

import (
	"time"
)

// FileContent представляет метаданные и важную информацию о файле, хранящемся в базе данных.
type FileContent struct {
	ID        string    `db:"id"`
	UserID    int64     `db:"user_id"`
	Name      []byte    `db:"name"`
	FilePath  string    `db:"file_path"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetID возвращает идентификатор (ID) объекта FileContent.
func (p FileContent) GetID() string {
	return p.ID
}

// FileWithComment — это структура, представляющая метаданные файла вместе со связанным комментарием, хранящимся в базе данных.
type FileWithComment struct {
	FileContent
	Comment string `db:"comment"`
}
