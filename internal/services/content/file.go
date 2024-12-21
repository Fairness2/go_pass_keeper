package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"passkeeper/internal/commonerrors"
	"passkeeper/internal/helpers"
	"passkeeper/internal/logger"
	"passkeeper/internal/models"
	"passkeeper/internal/payloads"
	"passkeeper/internal/repositories"
	"passkeeper/internal/token"
	"strconv"
	"time"
)

var (
	ErrEmptyName = errors.New("name is empty")
)

// FileService предоставляет методы для управления файлами пользователей и соответствующими комментариями в системе.
type FileService struct {
	dbPool   repositories.SQLExecutor
	filePath string
}

// NewFileService инициализирует и возвращает новый экземпляр FileService, настроенный с использованием предоставленной базой.
func NewFileService(dbPool repositories.SQLExecutor, filePath string) *FileService {
	return &FileService{
		dbPool:   dbPool,
		filePath: filePath,
	}
}

// SaveFileHandler обрабатывает HTTP-запросы на сохранение нового файла вместе с дополнительным комментарием для аутентифицированного пользователя.
// Он гарантирует корректность тела запроса, предотвращает предоставление идентификатора и связывает пароль с пользователем.
func (s *FileService) SaveFileHandler(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	// Читаем тело запроса
	fileBody, commentBody, err := s.getSaveFileBody(request, user.ID)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}
	repository := repositories.NewFileRepository(request.Context(), s.dbPool)
	if err = repository.Create(fileBody, commentBody); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
		if s.deleteFile(fileBody.FilePath, user.ID) != nil {
			logger.Log.Error(err)
		}
	}
}

// getSaveFileBody анализирует и проверяет тело HTTP-запроса для извлечения полезных данных или возвращает ошибку.
func (s *FileService) getSaveFileBody(request *http.Request, userID int64) (models.FileContent, models.Comment, error) {

	err := request.ParseMultipartForm(10 << 20)
	if err != nil {
		return models.FileContent{}, models.Comment{}, &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}
	file, _, err := request.FormFile("file")
	if err != nil {
		return models.FileContent{}, models.Comment{}, &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}
	defer file.Close()
	fileName := request.PostFormValue("name")
	if fileName == "" {
		return models.FileContent{}, models.Comment{}, &commonerrors.RequestError{InternalError: errors.New("name is empty"), HTTPStatus: http.StatusBadRequest}
	}

	filePath, err := s.saveFile(file, userID)
	if err != nil {
		return models.FileContent{}, models.Comment{}, &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusInternalServerError}
	}

	fileBody := models.FileContent{
		UserID:   userID,
		Name:     []byte(fileName),
		FilePath: filePath,
	}
	comment := models.Comment{
		ContentType: models.TypeFile,
		Comment:     request.PostFormValue("comment"),
	}
	return fileBody, comment, nil
}

// saveFile сохраняет содержимое данного файла на диск под уникальным именем в заданном пользователем каталоге.
// Возвращает сохраненный путь к файлу или ошибку, если операция не удалась.
func (s *FileService) saveFile(file io.Reader, userID int64) (string, error) {
	newFileName := uuid.New().String()
	dir := fmt.Sprintf("%s/%d", s.filePath, userID)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	filePath := fmt.Sprintf("%s/%s", dir, newFileName)
	dest, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer dest.Close()
	_, err = io.Copy(dest, file)
	if err != nil {
		return "", err
	}
	return newFileName, nil
}

// deleteFile удаляет файл из каталога определенного пользователя, создавая путь к файлу и вызывая os.Remove.
// Возвращает ошибку, если удаление не удалось (исключая случай, когда файл не существует).
func (s *FileService) deleteFile(filePath string, userID int64) error {
	newFilePath := fmt.Sprintf("%s/%d/%s", s.filePath, userID, filePath)
	if err := os.Remove(newFilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

// UpdateFileHandler обрабатывает HTTP-запросы на обновление существующего файла и связанного с ним комментария для аутентифицированного пользователя.
// Он проверяет тело запроса, проверяет наличие пароля пользователя и обеспечивает правильную обработку обновлений.
func (s *FileService) UpdateFileHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	body, err := s.getUpdateFileBody(request)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}
	// Для создания нельзя передавать идентификатор
	if body.ID <= 0 {
		helpers.ProcessResponseWithStatus("ID should not be empty", http.StatusBadRequest, response)
		return
	}
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	info := models.FileContent{
		ID:        body.ID,
		UserID:    user.ID,
		Name:      body.Name,
		UpdatedAt: time.Now(),
	}
	comment := models.Comment{
		ContentType: models.TypeFile,
		Comment:     body.Comment,
		ContentID:   body.ID,
		UpdatedAt:   time.Now(),
	}
	repository := repositories.NewFileRepository(request.Context(), s.dbPool)

	// Проверяем есть ли такой пароль у пользователя
	_, err = repository.GetByUserIDAndId(user.ID, info.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			helpers.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
			return
		} else {
			helpers.SetInternalError(err, response)
			return
		}
	}
	// Обновляем пароль
	if err = repository.Create(info, comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// getUpdateFileBody анализирует и проверяет тело HTTP-запроса для извлечения полезных данных обновления информации о файле или возвращает ошибку.
func (s *FileService) getUpdateFileBody(request *http.Request) (*payloads.UpdateFile, error) {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	// Парсим тело в структуру запроса
	var body payloads.UpdateFile
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}

	return &body, nil
}

// GetUserFiles обрабатывает запросы HTTP GET для получения всех файлов и комментариев к ним для аутентифицированного пользователя.
// Он извлекает пользователя из контекста запроса, извлекает соответствующие тексты из репозитория и возвращает их в формате JSON.
func (s *FileService) GetUserFiles(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	repository := repositories.NewFileRepository(request.Context(), s.dbPool)
	files, err := repository.GetByUserID(user.ID)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	filesData := make([]payloads.FileWithComment, 0, len(files))
	for _, file := range files {
		filesData = append(filesData, payloads.FileWithComment{
			ID:      file.ID,
			Name:    file.Name,
			Comment: file.Comment,
		})
	}
	marshaledBody, err := json.Marshal(filesData)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	if err = helpers.SetHTTPResponse(response, http.StatusOK, marshaledBody); err != nil {
		logger.Log.Error(err)
	}
}

// DeleteUserFile обрабатывает удаление записи файла пользователя по его идентификатору.
// Проверяет аутентификацию пользователя и гарантирует, что идентификатор пароля правильно анализируется из запроса.
func (s *FileService) DeleteUserFile(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	strID := chi.URLParam(request, "id")
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		helpers.ProcessResponseWithStatus("File ID is not correct", http.StatusBadRequest, response)
		return
	}
	repository := repositories.NewFileRepository(request.Context(), s.dbPool)
	file, err := repository.GetByUserIDAndId(user.ID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			helpers.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
			return
		} else {
			helpers.SetInternalError(err, response)
			return
		}
	}
	if err = s.deleteFile(file.FilePath, user.ID); err != nil {
		helpers.ProcessResponseWithStatus("Can`t delete file", http.StatusInternalServerError, response)
		return
	}

	if err = repository.DeleteByUserIDAndID(user.ID, id); err != nil {
		helpers.ProcessResponseWithStatus("Can`t delete", http.StatusInternalServerError, response)
		return
	}
}

func (s *FileService) DownloadFileHandler(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	strID := chi.URLParam(request, "id")
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		helpers.ProcessResponseWithStatus("File ID is not correct", http.StatusBadRequest, response)
		return
	}
	repository := repositories.NewFileRepository(request.Context(), s.dbPool)
	file, err := repository.GetByUserIDAndId(user.ID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			helpers.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
			return
		} else {
			helpers.SetInternalError(err, response)
			return
		}
	}
	filePath := fmt.Sprintf("%s/%d/%s", s.filePath, user.ID, file.FilePath)
	fileContent, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			helpers.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
			return
		}
		helpers.SetInternalError(err, response)
		return
	}
	fileStat, err := fileContent.Stat()
	if err != nil {
		helpers.SetInternalError(err, response)
	}
	response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.FilePath))
	response.Header().Set("Content-Type", "application/octet-stream")
	response.Header().Set("Content-Length", fmt.Sprintf("%d", fileStat.Size()))
	response.Header().Set("Content-Transfer-Encoding", "binary")
	response.Header().Set("Last-Modified", fileStat.ModTime().Format(http.TimeFormat))
	response.Header().Set("Accept-Ranges", "bytes")
	response.WriteHeader(http.StatusOK)

	// Write the file to the response
	if _, err = io.Copy(response, fileContent); err != nil {
		helpers.SetInternalError(err, response)
		return
	}
}
