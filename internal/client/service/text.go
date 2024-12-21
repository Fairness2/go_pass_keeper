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

type TextService struct {
	client *serverclient.Client
	user   *user.User
}

func NewTextService(client *serverclient.Client, user *user.User) *TextService {
	return &TextService{
		client: client,
		user:   user,
	}
}

func (s *TextService) GetTexts() ([]TextData, error) {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Get(textURL)
	if err != nil {
		return nil, errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return nil, ErrInvalidResponseStatus
	}
	var texts []payloads.TextWithComment
	err = json.Unmarshal(response.Body(), &texts)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	dTexts, err := s.DecryptTexts(texts)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	return dTexts, nil
}

func (s *TextService) EncryptText(body *payloads.TextWithComment) (*payloads.TextWithComment, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	if err := body.Encrypt(ch); err != nil {
		return body, err
	}
	return body, nil
}

func (s *TextService) DecryptTexts(texts []payloads.TextWithComment) ([]TextData, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	dTexts := make([]TextData, len(texts))
	for i, text := range texts {
		if err := text.Decrypt(ch); err != nil {
			return nil, err
		}
		dTexts[i] = TextData{
			TextWithComment: text,
			IsDecrypted:     true,
		}
	}

	return dTexts, nil
}

func (s *TextService) CreateText(body *payloads.TextWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Post(textURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *TextService) UpdateText(body *payloads.TextWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Put(textURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *TextService) DeleteText(id int64) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Delete(fmt.Sprintf("%s/%d", textURL, id))
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
