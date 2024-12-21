package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"passkeeper/internal/client/serverclient"
	"passkeeper/internal/client/user"
	"passkeeper/internal/encrypt/cipher"
	"passkeeper/internal/payloads"
)

const (
	fileURL = "/api/content/file"
)

var (
	ErrReadingFile    = errors.New("error reading file")
	ErrCreateTempFile = errors.New("error creating temp file for encrypting file")
	ErrEncryptingFile = errors.New("error encrypting file")
	ErrDecryptFile    = errors.New("error decrypt file")
)

// FileData представляет собой структуру данных карты с дополнительным комментарием и расшифрованным состоянием.
type FileData struct {
	payloads.FileWithComment
	IsDecrypted bool
	FilePath    string
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

func (s *FileService) EncryptFileInfo(body *payloads.FileWithComment) (*payloads.FileWithComment, error) {
	// TODO
	ch := cipher.NewCipher([]byte(s.user.Password))
	eName, err := ch.Encrypt(body.Name)
	if err != nil {
		return body, err
	}
	body.Name = eName

	return body, nil
}

func (s *FileService) EncryptFile(filePath string) (string, error) {
	originalFile, err := os.Open(filePath)
	if err != nil {
		return "", errors.Join(ErrReadingFile, err)
	}
	defer originalFile.Close()
	originalBody, err := io.ReadAll(originalFile)
	if err != nil {
		return "", errors.Join(ErrReadingFile, err)
	}
	encryptedFile, err := os.CreateTemp(os.TempDir(), "enc_*")
	if err != nil {
		return "", errors.Join(ErrCreateTempFile, err)
	}
	defer encryptedFile.Close()

	ch := cipher.NewCipher([]byte(s.user.Password))
	encryptedBody, err := ch.Encrypt(originalBody)
	if err != nil {
		return "", errors.Join(ErrEncryptingFile, err)
	}
	if _, err = encryptedFile.Write(encryptedBody); err != nil {
		return "", errors.Join(ErrEncryptingFile, err)
	}

	return encryptedFile.Name(), nil
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

func (s *FileService) DecryptFile(from io.Reader, dest io.Writer) error {
	enBody, err := io.ReadAll(from)
	if err != nil {
		return errors.Join(ErrReadingFile, err)
	}
	ch := cipher.NewCipher([]byte(s.user.Password))
	decryptedBody, err := ch.Decrypt(enBody)
	if err != nil {
		return errors.Join(ErrDecryptFile, err)
	}
	if _, err = dest.Write(decryptedBody); err != nil {
		return errors.Join(ErrDecryptFile, err)
	}
	return nil
}

func (s *FileService) CreateFile(body *payloads.FileWithComment, filePath string) error {
	request := s.client.Client.R()
	response, err := request.SetAuthToken(s.client.Token).
		SetFile("file", filePath).
		SetMultipartFormData(map[string]string{
			"name":    string(body.Name),
			"comment": body.Comment,
		}).
		Post(fileURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *FileService) UpdateFile(body *payloads.FileWithComment) error {
	response, err := s.client.Client.R().
		SetAuthToken(s.client.Token).
		SetBody(body).
		Put(fileURL)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *FileService) DeleteFile(id int64) error {
	response, err := s.client.Client.R().
		SetAuthToken(s.client.Token).
		Delete(fmt.Sprintf("%s/%d", fileURL, id))
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

func (s *FileService) DownloadFile(id int64, destFile string) error {
	req := s.client.Client.R().
		SetAuthToken(s.client.Token).
		SetOutput(destFile)
	resp, err := req.Get(fmt.Sprintf("%s/download/%d", fileURL, id))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
