package service

import (
	"passkeeper/internal/client/serverclient"
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

func NewCRUDTextService(client *serverclient.Client, user *user.User) *CRUDService[*payloads.TextWithComment, TextData] {
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
