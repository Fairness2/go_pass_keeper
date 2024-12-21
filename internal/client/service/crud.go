package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"passkeeper/internal/client/serverclient"
	"passkeeper/internal/client/user"
	"passkeeper/internal/encrypt/cipher"
	"passkeeper/internal/payloads"
)

type Encryptable interface {
	Encrypt(*cipher.Cipher) error
	Decrypt(*cipher.Cipher) error
}

type CRUDService[T Encryptable, Y list.Item] struct {
	client *serverclient.Client
	user   *user.User
	url    string
	crtY   func(T) Y
}

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

func (s *CRUDService[T, Y]) Get() ([]Y, error) {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Get(s.url)
	if err != nil {
		return nil, errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return nil, ErrInvalidResponseStatus
	}
	var items []T
	err = json.Unmarshal(response.Body(), &items)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	dItems, err := s.DecryptItems(items)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	return dItems, nil
}

func (s *CRUDService[T, Y]) EncryptItem(body T) (T, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	if err := body.Encrypt(ch); err != nil {
		return body, err
	}
	return body, nil
}

func (s *CRUDService[T, Y]) DecryptItems(items []T) ([]Y, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	dItems := make([]Y, len(items))
	for i, item := range items {
		if err := item.Decrypt(ch); err != nil {
			return nil, err
		}
		dItems[i] = s.crtY(item)
	}

	return dItems, nil
}

func (s *CRUDService[T, Y]) Create(body T) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Post(s.url)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *CRUDService[T, Y]) Update(body T) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Put(s.url)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *CRUDService[T, Y]) Delete(id int64) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Delete(fmt.Sprintf("%s/%d", s.url, id))
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
