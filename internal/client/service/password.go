package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"passkeeper/internal/client/serverclient"
	"passkeeper/internal/client/user"
	"passkeeper/internal/encrypt/cipher"
	"passkeeper/internal/payloads"
)

const (
	passURL = "/api/content/password"
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

func (s *PasswordService) GetPasswords() ([]payloads.PasswordWithComment, error) {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Get(passURL)
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

func (s *PasswordService) EncryptPassword(body *payloads.PasswordWithComment) (*payloads.PasswordWithComment, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	eUsername, err := ch.Encrypt(body.Username)
	if err != nil {
		return body, err
	}
	body.Username = eUsername
	ePass, err := ch.Encrypt(body.Password.Password)
	if err != nil {
		return body, err
	}
	body.Password.Password = ePass
	return body, nil
}

func (s *PasswordService) DecryptPassword(body *payloads.PasswordWithComment) (*payloads.PasswordWithComment, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	dUsername, err := ch.Decrypt(body.Username)
	if err != nil {
		return body, err
	}
	body.Username = dUsername
	dPass, err := ch.Decrypt(body.Password.Password)
	if err != nil {
		return body, err
	}
	body.Password.Password = dPass
	return body, nil
}

func (s *PasswordService) CreatePassword(body *payloads.PasswordWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Post(passURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *PasswordService) UpdatePassword(body *payloads.PasswordWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Put(passURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *PasswordService) DeletePassword(id int64) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Delete(fmt.Sprintf("%s/%d", passURL, id))
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
