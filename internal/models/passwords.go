package models

import (
	"time"
)

// PasswordContent представляет данные пароля пользователя, включая связанные метаданные, такие как домен, имя пользователя и временные метки.
type PasswordContent struct {
	ID        string    `db:"id"`
	UserID    int64     `db:"user_id"`
	Domen     string    `db:"domen"`
	Username  []byte    `db:"username"`
	Password  []byte    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// GetID возвращает уникальный идентификатор (ID) экземпляра PasswordContent.
func (p PasswordContent) GetID() string {
	return p.ID
}

// PasswordWithComment представляет собой запись пароля со связанным комментарием для дополнительного контекста или аннотации.
type PasswordWithComment struct {
	PasswordContent
	Comment string `db:"comment"`
}
