package models

import "time"

const (
	TypePassword = iota
	TypeFile
)

type ContentType = int8

type Comment struct {
	ID          int64       `db:"id"`
	ContentType ContentType `db:"content_type"`
	ContentID   int64       `db:"content_id"`
	Comment     string      `db:"comment"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}
