package models

import (
	"time"
)

// CardContent представляет собой запись информации о карте, хранящуюся в базе данных.
// Он включает поля для сведений о карте и связанных с ними метаданных, таких как отметки времени создания и обновления.
type CardContent struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Number    []byte    `db:"number"`
	Date      []byte    `db:"date"`
	Owner     []byte    `db:"owner"`
	CVV       []byte    `db:"cvv"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetID возвращает идентификатор экземпляра CardContent в формате int64.
func (p CardContent) GetID() int64 {
	return p.ID
}

// CardWithComment представляет запись карты вместе с соответствующим комментарием, хранящимся в базе данных.
type CardWithComment struct {
	CardContent
	Comment string `db:"comment"`
}
