package models

import (
	"time"
)

type TextContent struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	TextData  []byte    `db:"text_data"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type TextWithComment struct {
	TextContent
	Comment string `db:"comment"`
}
