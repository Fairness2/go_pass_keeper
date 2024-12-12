package service

import (
	"encoding/json"
	"errors"
	"passkeeper/internal/client/serverclient"
	"passkeeper/internal/payloads"
)

var (
	ErrEmptyUsername         = errors.New("username is empty")
	ErrEmptyPassword         = errors.New("password is empty")
	ErrInvalidRequestBody    = errors.New("invalid request body")
	ErrSendingRequest        = errors.New("error while sending request")
	ErrInvalidResponseStatus = errors.New("invalid response status")
	ErrInvalidResponseBody   = errors.New("invalid response body")
)

type LoginService struct {
	client *serverclient.Client
}

func NewLoginService(client *serverclient.Client) *LoginService {
	return &LoginService{client: client}
}

func (s *LoginService) Login(username, password string) error {
	if username == "" {
		return ErrEmptyUsername
	}
	if password == "" {
		return ErrEmptyPassword
	}
	body := payloads.Register{
		Login:    username,
		Password: password,
	}
	marshaledBody, err := json.Marshal(body)
	if err != nil {
		return errors.Join(ErrInvalidRequestBody, err)
	}
	request := s.client.Client.R()
	request.SetBody(marshaledBody)
	response, err := request.Post("/api/user/login")
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	var tokens payloads.Authorization
	err = json.Unmarshal(response.Body(), &tokens)
	if err != nil {
		return errors.Join(ErrInvalidResponseBody, err)
	}
	s.client.SetTokens(tokens.Token, tokens.Refresh)
	return nil
}
