package service

import (
	"passkeeper/internal/client/user"
	"passkeeper/internal/payloads"
)

const (
	textURL = "/api/content/text"
)

// TextData представляет собой структуру данных текста с дополнительным комментарием и расшифрованным состоянием.
type TextData struct {
	payloads.TextWithComment
	IsDecrypted bool
}

func (i TextData) Title() string {
	if !i.IsDecrypted {
		return "Не расшифровано"
	}
	text := []rune(string(i.TextData))
	if len(text) > 20 {
		return string(text[:17]) + "..."
	}

	return string(i.TextData)
}
func (i TextData) Description() string { return i.Comment }
func (i TextData) FilterValue() string { return string(i.TextData) }

// NewCRUDTextService создает новую службу CRUD для управления объектами TextWithComment.
// Он использует предоставленный серверный клиент и пользователя для запросов API.
// Служба включает в себя логику обработки и дешифрования текста для сущностей.
func NewCRUDTextService(client crudClient, user *user.User) *CRUDService[*payloads.TextWithComment, TextData] {
	return &CRUDService[*payloads.TextWithComment, TextData]{
		client: client,
		user:   user,
		url:    textURL,
		crtY: func(t *payloads.TextWithComment) TextData {
			return TextData{
				TextWithComment: *t,
				IsDecrypted:     true,
			}
		},
	}
}
