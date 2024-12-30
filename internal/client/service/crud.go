package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"github.com/go-resty/resty/v2"
	"passkeeper/internal/client/user"
	"passkeeper/internal/payloads"
)

// Encryptable определяет методы шифрования и дешифрования данных с использованием предоставленного шифра.
// Объекты, реализующие этот интерфейс, могут безопасно трансформировать свое внутреннее состояние.
type Encryptable interface {
	Encrypt(payloads.Encrypter) error
	Decrypt(payloads.Decrypter) error
}

// crudClient определяет интерфейс для взаимодействия с операциями CRUD на стороне сервера.
// GetRequest извлекает новый объект запроса, настроенный для клиента.
// GetToken возвращает токен аутентификации, связанный с клиентом.
type crudClient interface {
	GetRequest() *resty.Request
	GetToken() string
}

// CRUDService предоставляет базовые операции CRUD для зашифрованных данных с использованием клиента сервера и учетных данных пользователя.
// T — тип Encryptable, реализующий методы шифрования/дешифрования, а Y — преобразованный тип результата.
type CRUDService[T Encryptable, Y list.Item] struct {
	client crudClient
	user   *user.User
	url    string
	crtY   func(T) Y
}

// Get получает с сервера список расшифрованных элементов типа Y и возвращает их, либо ошибку в случае неудачи.
func (s *CRUDService[T, Y]) Get() ([]Y, error) {
	request := s.client.GetRequest()
	request.SetAuthToken(s.client.GetToken())
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

// EncryptItem шифрует данный элемент типа T, используя пароль пользователя в качестве ключа шифрования, и возвращает зашифрованный элемент.
func (s *CRUDService[T, Y]) EncryptItem(body T) (T, error) {
	ch := s.user.Cipher
	if err := body.Encrypt(ch); err != nil {
		return body, err
	}
	return body, nil
}

// DecryptItems расшифровывает фрагмент элементов типа T в фрагмент типа Y, используя пароль пользователя в качестве ключа дешифрования.
// Возвращает расшифрованный фрагмент или ошибку, если расшифровка не удалась на каком-либо этапе.
func (s *CRUDService[T, Y]) DecryptItems(items []T) ([]Y, error) {
	ch := s.user.Cipher
	dItems := make([]Y, len(items))
	for i, item := range items {
		if err := item.Decrypt(ch); err != nil {
			return nil, err
		}
		dItems[i] = s.crtY(item)
	}

	return dItems, nil
}

// Create отправляет POST-запрос на создание нового ресурса типа T и возвращает ошибку, если запрос не выполнен.
func (s *CRUDService[T, Y]) Create(body T) error {
	request := s.client.GetRequest()
	request.SetAuthToken(s.client.GetToken())
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

// Update отправляет запрос PUT с предоставленным телом типа T для обновления существующего ресурса и возвращает ошибку в случае сбоя.
func (s *CRUDService[T, Y]) Update(body T) error {
	request := s.client.GetRequest()
	request.SetAuthToken(s.client.GetToken())
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

// Delete отправляет запрос DELETE на удаление ресурса по его идентификатору и возвращает ошибку, если операция не удалась.
func (s *CRUDService[T, Y]) Delete(id string) error {
	request := s.client.GetRequest()
	request.SetAuthToken(s.client.GetToken())
	response, err := request.Delete(fmt.Sprintf("%s/%s", s.url, id))
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
