package models

import (
	"time"
)

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

type CardWithComment struct {
	CardContent
	Comment string `db:"comment"`
}
