package content

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"passkeeper/internal/commonerrors"
	"passkeeper/internal/logger"
	"passkeeper/internal/models"
	"passkeeper/internal/payloads"
	"passkeeper/internal/repositories"
	"passkeeper/internal/responsesetters"
	"passkeeper/internal/token"
	"time"
)

var (
	ErrEmptyName = errors.New("name is empty")
)

// fileRepository определяет методы взаимодействия с файлами и соответствующими комментариями в системе.
type fileRepository interface {
	Create(ctx context.Context, content models.FileContent, comment models.Comment) error
	GetByUserIDAndId(ctx context.Context, userID int64, id string) (*models.FileContent, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.FileWithComment, error)
	DeleteByUserIDAndID(ctx context.Context, userID int64, id string) error
}

// FileService предоставляет методы для управления файлами пользователей и соответствующими комментариями в системе.
type FileService struct {
	repository  fileRepository
	filePath    string
	permissions os.FileMode
	maxFormSize int64
}

// NewFileService инициализирует и возвращает новый экземпляр FileService, настроенный с использованием предоставленной базой.
func NewFileService(rep fileRepository, filePath string) *FileService {
	return &FileService{
		repository:  rep,
		filePath:    filePath,
		permissions: os.ModePerm,
		maxFormSize: 10 << 20,
	}
}

// SaveFileHandler обрабатывает HTTP-запросы на сохранение нового файла вместе с дополнительным комментарием для аутентифицированного пользователя.
// Он гарантирует корректность тела запроса, предотвращает предоставление идентификатора и связывает пароль с пользователем.
//
//	@Summary		Сохраните новый файл с необязательным комментарием.
//	@Description	Сохраняет новый файл для аутентифицированного пользователя. Запрос должен включать составной файл, имя и необязательный комментарий.
//	@Tags			file
//	@Accept			multipart/form-data
//	@Produce		application/json
//	@Param			file	formData	file	true	"File to upload"
//	@Param			name	formData	string	true	"Name of the file"
//	@Param			comment	formData	string	false	"Optional comment for the file"
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid input or bad request"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/files [post]
func (s *FileService) SaveFileHandler(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	// Читаем тело запроса
	fileBody, commentBody, err := s.getSaveFileBody(request, user.ID)
	if err != nil {
		responsesetters.ProcessRequestErrorWithBody(err, response)
		return
	}
	if err = s.repository.Create(request.Context(), fileBody, commentBody); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
		if s.deleteFile(fileBody.FilePath, user.ID) != nil {
			logger.Log.Error(err)
		}
	}
}

// getSaveFileBody анализирует и проверяет тело HTTP-запроса для извлечения полезных данных или возвращает ошибку.
func (s *FileService) getSaveFileBody(request *http.Request, userID int64) (models.FileContent, models.Comment, error) {
	err := request.ParseMultipartForm(s.maxFormSize)
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
	if err := os.MkdirAll(dir, s.permissions); err != nil {
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
//
//	@Summary		Обновить файл пользователя с комментарием
//	@Description	Обновляет существующий файл вместе с комментарием для аутентифицированного пользователя. Гарантирует существование файла и обрабатывает проверку.
//	@Tags			files
//	@Accept			json
//	@Produce		json
//	@Param			data	body	payloads.UpdateFile	true	"Updated file data and comment"
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid input or bad request"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		404	{object}	payloads.ErrorResponseBody	"File not found"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/file [put]
func (s *FileService) UpdateFileHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	var body payloads.UpdateFile
	err := getSaveBody(request, &body)
	if err != nil {
		responsesetters.ProcessRequestErrorWithBody(err, response)
		return
	}
	// Для создания нельзя передавать идентификатор
	if body.ID == "" {
		responsesetters.ProcessResponseWithStatus("ID should not be empty", http.StatusBadRequest, response)
		return
	}
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
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
	ctx := request.Context()
	// Проверяем есть ли такой пароль у пользователя
	_, err = s.repository.GetByUserIDAndId(ctx, user.ID, info.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			responsesetters.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
			return
		} else {
			responsesetters.SetInternalError(err, response)
			return
		}
	}
	// Обновляем пароль
	if err = s.repository.Create(ctx, info, comment); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// GetUserFiles обрабатывает запросы HTTP GET для получения всех файлов и комментариев к ним для аутентифицированного пользователя.
// Он извлекает пользователя из контекста запроса, извлекает соответствующие тексты из репозитория и возвращает их в формате JSON.
//
//	@Summary		Получить файлы пользователей с комментариями
//	@Description	Получает список файлов и связанных с ними комментариев для аутентифицированного пользователя.
//	@Tags			file
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		payloads.FileWithComment	"Array of user files with comments"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authorized"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/file [get]
func (s *FileService) GetUserFiles(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	files, err := s.repository.GetByUserID(request.Context(), user.ID)
	if err != nil {
		responsesetters.SetInternalError(err, response)
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
		responsesetters.SetInternalError(err, response)
		return
	}
	if err = responsesetters.SetHTTPResponse(response, http.StatusOK, marshaledBody); err != nil {
		logger.Log.Error(err)
	}
}

// DeleteUserFile обрабатывает удаление записи файла пользователя по его идентификатору.
// Проверяет аутентификацию пользователя и гарантирует, что идентификатор пароля правильно анализируется из запроса.
//
//	@Summary		Удалить файл пользователя
//	@Description	Удаляет файл пользователя по его идентификатору. Гарантирует, что пользователь аутентифицирован и идентификатор файла корректен.
//	@Tags			file
//	@Param			id	path	string	true	"File ID"
//	@Produce		json
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid file ID"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/file/{id} [delete]
func (s *FileService) DeleteUserFile(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	id, err := getIDFromRequest(request)
	if err != nil {
		responsesetters.ProcessRequestErrorWithBody(err, response)
		return
	}
	ctx := request.Context()
	file, err := s.repository.GetByUserIDAndId(ctx, user.ID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			responsesetters.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
		} else {
			responsesetters.SetInternalError(err, response)
		}
		return
	}
	if err = s.deleteFile(file.FilePath, user.ID); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t delete file", http.StatusInternalServerError, response)
		return
	}

	if err = s.repository.DeleteByUserIDAndID(ctx, user.ID, id); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t delete", http.StatusInternalServerError, response)
		return
	}
}

// DownloadFileHandler обрабатывает запросы на загрузку файлов для авторизованных пользователей, получая и предоставляя запрошенный файл.
//
//	@Summary		Загрузите файл пользователя
//	@Description	Позволяет аутентифицированному пользователю загружать файл по его идентификатору.
//	@Tags			file
//	@Param			id	path	string	true	"File ID"
//	@Produce		octet-stream
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid file ID"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		404	{object}	payloads.ErrorResponseBody	"File not found"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/file/download/{id} [get]
func (s *FileService) DownloadFileHandler(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	id, err := getIDFromRequest(request)
	if err != nil {
		responsesetters.ProcessRequestErrorWithBody(err, response)
		return
	}
	file, err := s.repository.GetByUserIDAndId(request.Context(), user.ID, id)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			responsesetters.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
			return
		} else {
			responsesetters.SetInternalError(err, response)
			return
		}
	}
	filePath := fmt.Sprintf("%s/%d/%s", s.filePath, user.ID, file.FilePath)
	fileContent, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			responsesetters.ProcessResponseWithStatus("File not found", http.StatusNotFound, response)
			return
		}
		responsesetters.SetInternalError(err, response)
		return
	}
	fileStat, err := fileContent.Stat()
	if err != nil {
		responsesetters.SetInternalError(err, response)
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
		responsesetters.SetInternalError(err, response)
		return
	}
}

// RegisterRoutes настраивает HTTP-маршруты для FileService, применяя предоставленное промежуточное программное обеспечение для обработки файлов.
func (s *FileService) RegisterRoutes(middlewares ...func(http.Handler) http.Handler) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/file", s.SaveFileHandler)
		r.Put("/file", s.UpdateFileHandler)
		r.Get("/file", s.GetUserFiles)
		r.Delete("/file/{id}", s.DeleteUserFile)
		r.Get("/file/download/{id}", s.DownloadFileHandler)
	}
}
