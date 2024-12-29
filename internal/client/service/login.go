package service

import (
	"encoding/json"
	"errors"
	"github.com/go-resty/resty/v2"
	"passkeeper/internal/client/user"
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

// loginService определяет методы обработки запросов аутентификации и управления токенами аутентификации.
type loginService interface {
	GetRequest() *resty.Request
	SetTokens(token, refreshToken string)
}

// LoginService предоставляет методы аутентификации пользователей с использованием клиента для связи с сервером.
type LoginService struct {
	client loginService
}

// NewLoginService инициализирует новый экземпляр LoginService, используя предоставленный серверный клиент для аутентификации пользователя.
func NewLoginService(client loginService) *LoginService {
	return &LoginService{client: client}
}

// Login аутентифицирует пользователя, отправляя его логин и пароль на сервер, и обрабатывает полученные токены.
func (s *LoginService) Login(username, password string, isRegistration bool) error {
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
	request := s.client.GetRequest()
	request.SetBody(marshaledBody)
	var response *resty.Response
	if !isRegistration {
		response, err = request.Post("/api/user/login")
	} else {
		response, err = request.Post("/api/user/register")
	}
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
	user.CurrentUser = &user.User{
		Password: password,
	}
	return nil
}
