package models

import (
	"time"
)

type FileContent struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Name      []byte    `db:"name"`
	FilePath  string    `db:"file_path"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type FileWithComment struct {
	FileContent
	Comment string `db:"comment"`
}
