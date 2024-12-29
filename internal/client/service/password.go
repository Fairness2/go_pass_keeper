package service

import (
	"passkeeper/internal/client/user"
	"passkeeper/internal/payloads"
)

const (
	passURL = "/api/content/password"
)

// PassData представляет собой структуру данных пароля с дополнительным комментарием и расшифрованным состоянием.
type PassData struct {
	payloads.PasswordWithComment
	isDecrypted bool
}

func (i PassData) Title() string       { return i.Domen }
func (i PassData) Description() string { return i.Comment }
func (i PassData) FilterValue() string { return i.Domen }

// NewCRUDPasswordService создает и возвращает новый CRUDService для управления PasswordWithComment с поддержкой расшифровки.
func NewCRUDPasswordService(client crudClient, user *user.User) *CRUDService[*payloads.PasswordWithComment, PassData] {
	return &CRUDService[*payloads.PasswordWithComment, PassData]{
		client: client,
		user:   user,
		url:    passURL,
		crtY: func(t *payloads.PasswordWithComment) PassData {
			return PassData{
				PasswordWithComment: *t,
				isDecrypted:         true,
			}
		},
	}
}
