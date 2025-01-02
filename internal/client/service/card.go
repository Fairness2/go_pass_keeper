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

// NewCRUDCardService инициализирует CRUDService для управления данными карты с параметрами аутентификации сервера-клиента и пользователя.
// Возвращает CRUDService, настроенный для обработки полезных данных CardWithComment и преобразованный в CardData.
func NewCRUDCardService(client crudClient, user *user.User) *CRUDService[*payloads.CardWithComment, CardData] {
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

// NewDefaultCardService создает службу CRUD по умолчанию для управления картами с использованием глобального клиента для взаимодействия с сервером и текущего пользователя.
func NewDefaultCardService() *CRUDService[*payloads.CardWithComment, CardData] {
	return NewCRUDCardService(serverclient.Inst, user.CurrentUser)
}
