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

type CardService struct {
	client *serverclient.Client
	user   *user.User
}

func NewCardService(client *serverclient.Client, user *user.User) *CardService {
	return &CardService{
		client: client,
		user:   user,
	}
}

func (s *CardService) GetCards() ([]CardData, error) {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Get(cardURL)
	if err != nil {
		return nil, errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return nil, ErrInvalidResponseStatus
	}
	var cards []payloads.CardWithComment
	err = json.Unmarshal(response.Body(), &cards)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	dCards, err := s.DecryptCards(cards)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	return dCards, nil
}

func (s *CardService) EncryptCard(body *payloads.CardWithComment) (*payloads.CardWithComment, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	eNumber, err := ch.Encrypt(body.Number)
	if err != nil {
		return body, err
	}
	body.Number = eNumber
	if len(body.Date) > 0 {
		eDate, err := ch.Encrypt(body.Date)
		if err != nil {
			return body, err
		}
		body.Date = eDate
	}
	if len(body.Owner) > 0 {
		eOwner, err := ch.Encrypt(body.Owner)
		if err != nil {
			return body, err
		}
		body.Owner = eOwner
	}
	if len(body.CVV) > 0 {
		eCVV, err := ch.Encrypt(body.CVV)
		if err != nil {
			return body, err
		}
		body.CVV = eCVV
	}
	return body, nil
}

func (s *CardService) DecryptCards(cards []payloads.CardWithComment) ([]CardData, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	dCards := make([]CardData, len(cards))
	for i, card := range cards {
		dNumber, err := ch.Decrypt(card.Number)
		if err != nil {
			return nil, err
		}
		card.Number = dNumber
		if len(card.Date) > 0 {
			eDate, err := ch.Decrypt(card.Date)
			if err != nil {
				return nil, err
			}
			card.Date = eDate
		}
		if len(card.Owner) > 0 {
			eOwner, err := ch.Decrypt(card.Owner)
			if err != nil {
				return nil, err
			}
			card.Owner = eOwner
		}
		if len(card.CVV) > 0 {
			eCVV, err := ch.Decrypt(card.CVV)
			if err != nil {
				return nil, err
			}
			card.CVV = eCVV
		}
		dCards[i] = CardData{
			CardWithComment: card,
			IsDecrypted:     true,
		}
	}

	return dCards, nil
}

func (s *CardService) CreateCard(body *payloads.CardWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Post(cardURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *CardService) UpdateCard(body *payloads.CardWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Put(cardURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *CardService) DeleteCard(id int64) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Delete(fmt.Sprintf("%s/%d", cardURL, id))
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
