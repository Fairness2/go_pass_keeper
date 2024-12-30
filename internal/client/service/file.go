package service

import (
	"errors"
	"fmt"
	"io"
	"os"
	"passkeeper/internal/client/user"
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

// FileService предоставляет функции для управления файлами и связанными с ними комментариями,
// использование CRUDService для стандартных операций, таких как создание, чтение, обновление и удаление.
type FileService struct {
	CRUDService[*payloads.FileWithComment, FileData]
}

// NewFileService создает и инициализирует новый экземпляр FileService с предоставленными конфигурациями клиента и пользователя.
func NewFileService(client crudClient, user *user.User) *FileService {
	return &FileService{
		CRUDService[*payloads.FileWithComment, FileData]{
			client: client,
			user:   user,
			url:    fileURL,
			crtY: func(t *payloads.FileWithComment) FileData {
				return FileData{
					FileWithComment: *t,
					IsDecrypted:     true,
				}
			},
		},
	}
}

// EncryptFile шифрует содержимое файла по заданному пути к файлу и создает временный зашифрованный файл.
// Возвращает путь к зашифрованному временному файлу или ошибку в случае сбоя процесса.
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

	ch := s.user.Cipher
	encryptedBody, err := ch.Encrypt(originalBody)
	if err != nil {
		return "", errors.Join(ErrEncryptingFile, err)
	}
	if _, err = encryptedFile.Write(encryptedBody); err != nil {
		return "", errors.Join(ErrEncryptingFile, err)
	}

	return encryptedFile.Name(), nil
}

// DecryptFile расшифровывает содержимое предоставленного io.Reader и записывает расшифрованные данные в указанный io.Writer.
// Возвращает ошибку, если чтение, расшифровка или запись не удались.
func (s *FileService) DecryptFile(from io.Reader, dest io.Writer) error {
	enBody, err := io.ReadAll(from)
	if err != nil {
		return errors.Join(ErrReadingFile, err)
	}
	ch := s.user.Cipher
	decryptedBody, err := ch.Decrypt(enBody)
	if err != nil {
		return errors.Join(ErrDecryptFile, err)
	}
	if _, err = dest.Write(decryptedBody); err != nil {
		return errors.Join(ErrDecryptFile, err)
	}
	return nil
}

// CreateFile отправляет файл и связанные с ним метаданные на сервер для создания.
// В качестве входных данных требуется объект FileWithComment и путь к файлу.
// Возвращает ошибку, если запрос не выполнен или статус ответа недействителен.
func (s *FileService) CreateFile(body *payloads.FileWithComment, filePath string) error {
	request := s.client.GetRequest()
	response, err := request.SetAuthToken(s.client.GetToken()).
		SetFile("file", filePath).
		SetMultipartFormData(map[string]string{
			"name":    string(body.Name),
			"comment": body.Comment,
		}).
		Post(s.url)
	if err != nil {
		return errors.Join(ErrSendingRequest, err)
	}
	if response.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}

// DownloadFile загружает файл с указанным идентификатором с сервера и сохраняет его по указанному пути назначения.
func (s *FileService) DownloadFile(id string, destFile string) error {
	req := s.client.GetRequest().
		SetAuthToken(s.client.GetToken()).
		SetOutput(destFile)
	resp, err := req.Get(fmt.Sprintf("%s/download/%s", s.url, id))
	if err != nil {
		return err
	}
	if resp.StatusCode() != 200 {
		return ErrInvalidResponseStatus
	}
	return nil
}
