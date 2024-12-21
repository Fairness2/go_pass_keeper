package content

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
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

// PasswordService предоставляет методы для управления паролями пользователей и соответствующими комментариями в системе.
type PasswordService struct {
	dbPool repositories.SQLExecutor
}

// NewPasswordService инициализирует и возвращает новый экземпляр PasswordService, настроенный с использованием предоставленной базой.
func NewPasswordService(dbPool repositories.SQLExecutor) *PasswordService {
	return &PasswordService{
		dbPool: dbPool,
	}
}

// SavePasswordHandler обрабатывает HTTP-запросы на сохранение нового пароля вместе с дополнительным комментарием для аутентифицированного пользователя.
// Он гарантирует корректность тела запроса, предотвращает предоставление идентификатора и связывает пароль с пользователем.
func (s *PasswordService) SavePasswordHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	body, err := s.getSavePasswordBody(request)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}
	// Для создания нельзя передавать идентификатор
	if body.ID != 0 {
		helpers.ProcessResponseWithStatus("ID should be empty", http.StatusBadRequest, response)
		return
	}

	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	pass := models.PasswordContent{
		UserID:   user.ID,
		Domen:    body.Domen,
		Username: body.Username,
		Password: body.Password,
	}
	comment := models.Comment{
		ContentType: models.TypePassword,
		Comment:     body.Comment,
	}
	repository := repositories.NewPasswordRepository(request.Context(), s.dbPool)
	if err = repository.Create(pass, comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// getSavePasswordBody анализирует и проверяет тело HTTP-запроса для извлечения полезных данных SavePassword или возвращает ошибку.
func (s *PasswordService) getSavePasswordBody(request *http.Request) (*payloads.SavePassword, error) {
	// TODO валидация
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	// Парсим тело в структуру запроса
	var body payloads.SavePassword
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}

	return &body, nil
}

// UpdatePasswordHandler обрабатывает HTTP-запросы на обновление существующего пароля и связанного с ним комментария для аутентифицированного пользователя.
// Он проверяет тело запроса, проверяет наличие пароля пользователя и обеспечивает правильную обработку обновлений.
func (s *PasswordService) UpdatePasswordHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	body, err := s.getSavePasswordBody(request)
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
	pass := models.PasswordContent{
		ID:        body.ID,
		UserID:    user.ID,
		Domen:     body.Domen,
		Username:  body.Username,
		Password:  body.Password,
		UpdatedAt: time.Now(),
	}
	comment := models.Comment{
		ContentType: models.TypePassword,
		Comment:     body.Comment,
		ContentID:   body.ID,
		UpdatedAt:   time.Now(),
	}
	repository := repositories.NewPasswordRepository(request.Context(), s.dbPool)

	// Проверяем есть ли такой пароль у пользователя
	_, err = repository.GetByUserIDAndId(pass.UserID, pass.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			helpers.ProcessResponseWithStatus("Password not found", http.StatusNotFound, response)
			return
		} else {
			helpers.SetInternalError(err, response)
			return
		}
	}
	// Обновляем пароль
	if err = repository.Create(pass, comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// GetUserPasswords обрабатывает запросы HTTP GET для получения всех паролей и комментариев к ним для аутентифицированного пользователя.
// Он извлекает пользователя из контекста запроса, извлекает соответствующие пароли из репозитория и возвращает их в формате JSON.
func (s *PasswordService) GetUserPasswords(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	repository := repositories.NewPasswordRepository(request.Context(), s.dbPool)
	passwords, err := repository.GetByUserID(user.ID)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	passwordsData := make([]payloads.PasswordWithComment, 0, len(passwords))
	for _, password := range passwords {
		passwordsData = append(passwordsData, payloads.PasswordWithComment{
			Password: payloads.Password{
				ID:       password.ID,
				Domen:    password.Domen,
				Username: password.Username,
				Password: password.Password,
			},
			Comment: password.Comment,
		})
	}
	marshaledBody, err := json.Marshal(passwordsData)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	if err = helpers.SetHTTPResponse(response, http.StatusOK, marshaledBody); err != nil {
		logger.Log.Error(err)
	}
}

// DeleteUserPasswords обрабатывает удаление записи пароля пользователя по его идентификатору.
// Проверяет аутентификацию пользователя и гарантирует, что идентификатор пароля правильно анализируется из запроса.
func (s *PasswordService) DeleteUserPasswords(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	strID := chi.URLParam(request, "id")
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		helpers.ProcessResponseWithStatus("Password ID is not correct", http.StatusBadRequest, response)
		return
	}
	repository := repositories.NewPasswordRepository(request.Context(), s.dbPool)
	if err = repository.DeleteByUserIDAndID(user.ID, id); err != nil {
		helpers.ProcessResponseWithStatus("Can`t delete", http.StatusInternalServerError, response)
		return
	}
}
