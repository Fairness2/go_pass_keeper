package service

import (
	"passkeeper/internal/client/serverclient"
	"passkeeper/internal/client/user"
	"passkeeper/internal/payloads"
)

const (
	cardURL = "/api/content/card"
)

// CardData представляет собой структуру данных карты с дополнительным комментарием и расшифрованным состоянием.
type CardData struct {
	payloads.CardWithComment
	IsDecrypted bool
}

func (i CardData) Title() string {
	if !i.IsDecrypted {
		return "Не расшифровано"
	}
	return string(i.Number)
}
func (i CardData) Description() string { return i.Comment }
func (i CardData) FilterValue() string { return string(i.Number) }

func NewCRUDCardService(client *serverclient.Client, user *user.User) *CRUDService[*payloads.CardWithComment, CardData] {
	return &CRUDService[*payloads.CardWithComment, CardData]{
		client: client,
		user:   user,
		url:    cardURL,
		crtY: func(t *payloads.CardWithComment) CardData {
			return CardData{
				CardWithComment: *t,
				IsDecrypted:     true,
			}
		},
	}
}
