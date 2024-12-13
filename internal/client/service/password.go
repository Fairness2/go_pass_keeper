package service

import (
	"encoding/json"
	"errors"
	"passkeeper/internal/client/serverclient"
	"passkeeper/internal/client/user"
	"passkeeper/internal/payloads"
)

type PasswordService struct {
	client *serverclient.Client
	user   *user.User
}

func NewPasswordService(client *serverclient.Client, user *user.User) *PasswordService {
	return &PasswordService{
		client: client,
		user:   user,
	}
}

func (s PasswordService) GetPasswords() ([]payloads.PasswordWithComment, error) {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Get("/api/content/password")
	if err != nil {
		return nil, errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return nil, ErrInvalidResponseStatus
	}
	var passwords []payloads.PasswordWithComment
	err = json.Unmarshal(response.Body(), &passwords)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	return passwords, err
}
