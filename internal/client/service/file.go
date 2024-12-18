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
	fileURL = "/api/content/file"
)

// FileData представляет собой структуру данных карты с дополнительным комментарием и расшифрованным состоянием.
type FileData struct {
	payloads.FileWithComment
	IsDecrypted bool
}

func (i FileData) Title() string {
	if !i.IsDecrypted {
		return "Не расшифровано"
	}
	return string(i.Name)
}
func (i FileData) Description() string { return i.Comment }
func (i FileData) FilterValue() string { return string(i.Name) }

type FileService struct {
	client *serverclient.Client
	user   *user.User
}

func NewFileService(client *serverclient.Client, user *user.User) *FileService {
	return &FileService{
		client: client,
		user:   user,
	}
}

func (s *FileService) GetFiles() ([]FileData, error) {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Get(fileURL)
	if err != nil {
		return nil, errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return nil, ErrInvalidResponseStatus
	}
	var files []payloads.FileWithComment
	err = json.Unmarshal(response.Body(), &files)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	dCards, err := s.DecryptFiles(files)
	if err != nil {
		return nil, errors.Join(ErrInvalidResponseBody, err)
	}
	return dCards, nil
}

func (s *FileService) EncryptFile(body *payloads.FileWithComment) (*payloads.FileWithComment, error) {
	// TODO
	ch := cipher.NewCipher([]byte(s.user.Password))
	eName, err := ch.Encrypt(body.Name)
	if err != nil {
		return body, err
	}
	body.Name = eName

	return body, nil
}

func (s *FileService) DecryptFiles(files []payloads.FileWithComment) ([]FileData, error) {
	ch := cipher.NewCipher([]byte(s.user.Password))
	dFiles := make([]FileData, len(files))
	for i, file := range files {
		dName, err := ch.Decrypt(file.Name)
		if err != nil {
			return nil, err
		}
		file.Name = dName
		dFiles[i] = FileData{
			FileWithComment: file,
			IsDecrypted:     true,
		}
	}

	return dFiles, nil
}

func (s *FileService) CreateFile(body *payloads.FileWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Post(fileURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *FileService) UpdateFile(body *payloads.FileWithComment) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	request.SetBody(body)
	response, err := request.Put(fileURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *FileService) DeleteFile(id int64) error {
	request := s.client.Client.R()
	request.SetAuthToken(s.client.Token)
	response, err := request.Delete(fmt.Sprintf("%s/%d", fileURL, id))
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
