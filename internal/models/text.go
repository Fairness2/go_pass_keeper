package models

import (
	"time"
)

// TextContent представляет запись текстового контента, включая метаданные, такие как ассоциация пользователя и временные метки.
type TextContent struct {
	ID        string    `db:"id"`
	UserID    int64     `db:"user_id"`
	TextData  []byte    `db:"text_data"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetID возвращает уникальный идентификатор (ID) экземпляра TextContent.
func (p TextContent) GetID() string {
	return p.ID
}

// TextWithComment представляет запись текстового содержимого вместе со связанным с ней комментарием. Он расширяет тип TextContent.
type TextWithComment struct {
	TextContent
	Comment string `db:"comment"`
}
