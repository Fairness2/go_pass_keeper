package models

import "time"

const (
	TypePassword ContentType = iota
	TypeFile
	TypeText
	TypeCard
)

// ContentType — это псевдоним типа для int8, используемый для представления различных типов контента в виде перечисляемых значений.
type ContentType = int8

// Comment представляет комментарий, связанный с определенным типом контента и идентификатором контента в системе.
type Comment struct {
	ID          int64       `db:"id"`
	ContentType ContentType `db:"content_type"`
	ContentID   int64       `db:"content_id"`
	Comment     string      `db:"comment"`
	CreatedAt   time.Time   `db:"created_at"`
	UpdatedAt   time.Time   `db:"updated_at"`
}
