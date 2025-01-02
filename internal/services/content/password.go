package content

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
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
	ErrBadRequest = errors.New("bad request")
)

// passwordRepository определяет интерфейс для управления паролями пользователей и соответствующими комментариями в системе.
type passwordRepository interface {
	Create(ctx context.Context, content models.PasswordContent, comment models.Comment) error
	GetByUserIDAndId(ctx context.Context, userID int64, id string) (*models.PasswordContent, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.PasswordWithComment, error)
	DeleteByUserIDAndID(ctx context.Context, userID int64, id string) error
}

// PasswordService предоставляет методы для управления паролями пользователей и соответствующими комментариями в системе.
type PasswordService struct {
	repository passwordRepository
}

// NewPasswordService инициализирует и возвращает новый экземпляр PasswordService, настроенный с использованием предоставленной базой.
func NewPasswordService(rep passwordRepository) *PasswordService {
	return &PasswordService{
		repository: rep,
	}
}

// SavePasswordHandler обрабатывает HTTP-запросы на сохранение нового пароля вместе с дополнительным комментарием для аутентифицированного пользователя.
// Он гарантирует корректность тела запроса, предотвращает предоставление идентификатора и связывает пароль с пользователем.
//
//	@Summary		Сохранить пользовательский пароль с комментарием
//	@Description	Создает новую запись о пароле вместе с необязательным комментарием для аутентифицированного пользователя.
//	@Tags			text
//	@Accept			json
//	@Produce		json
//	@Param			data	body	payloads.SavePassword	true	"Password data and comment"
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid input or bad request"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/password [post]
func (s *PasswordService) SavePasswordHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	var body payloads.SavePassword
	err := getSaveBody(request, &body)
	if err != nil {
		responsesetters.ProcessRequestErrorWithBody(err, response)
		return
	}

	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
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
	if err = s.repository.Create(request.Context(), pass, comment); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// getSaveBody анализирует и проверяет тело HTTP-запроса для извлечения полезных данных SavePassword или возвращает ошибку.
func getSaveBody(request *http.Request, body any) error {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		return err
	}
	// Парсим тело в структуру запроса
	err = json.Unmarshal(rawBody, body)
	if err != nil {
		return &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}

	result, err := govalidator.ValidateStruct(body)
	if !result {
		return &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}

	return nil
}

// UpdatePasswordHandler обрабатывает HTTP-запросы на обновление существующего пароля и связанного с ним комментария для аутентифицированного пользователя.
// Он проверяет тело запроса, проверяет наличие пароля пользователя и обеспечивает правильную обработку обновлений.
//
//	@Summary		Обновить пароль пользователя с комментарием
//	@Description	Обновляет существующий пароль вместе с комментарием для аутентифицированного пользователя. Гарантирует существование пароля и обрабатывает проверку.
//	@Tags			password
//	@Accept			json
//	@Produce		json
//	@Param			data	body	payloads.SavePassword	true	"Updated password data and comment"
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid input or bad request"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		404	{object}	payloads.ErrorResponseBody	"Password not found"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/password [put]
func (s *PasswordService) UpdatePasswordHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	var body payloads.UpdatePassword
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
	ctx := request.Context()
	// Проверяем есть ли такой пароль у пользователя
	_, err = s.repository.GetByUserIDAndId(ctx, pass.UserID, pass.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			responsesetters.ProcessResponseWithStatus("Password not found", http.StatusNotFound, response)
			return
		} else {
			responsesetters.SetInternalError(err, response)
			return
		}
	}
	// Обновляем пароль
	if err = s.repository.Create(ctx, pass, comment); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// GetUserPasswords обрабатывает запросы HTTP GET для получения всех паролей и комментариев к ним для аутентифицированного пользователя.
// Он извлекает пользователя из контекста запроса, извлекает соответствующие пароли из репозитория и возвращает их в формате JSON.
//
//	@Summary		Получить пароли пользователей с комментариями
//	@Description	Получает список паролей и связанных с ними комментариев для аутентифицированного пользователя.
//	@Tags			password
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		payloads.PasswordWithComment	"Array of user passwords with comments"
//	@Failure		401	{object}	payloads.ErrorResponseBody		"User not authorized"
//	@Failure		500	{object}	payloads.ErrorResponseBody		"Internal server error"
//	@Router			/api/content/password [get]
func (s *PasswordService) GetUserPasswords(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	passwords, err := s.repository.GetByUserID(request.Context(), user.ID)
	if err != nil {
		responsesetters.SetInternalError(err, response)
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
		responsesetters.SetInternalError(err, response)
		return
	}
	if err = responsesetters.SetHTTPResponse(response, http.StatusOK, marshaledBody); err != nil {
		logger.Log.Error(err)
	}
}

// DeleteUserPasswords обрабатывает удаление записи пароля пользователя по его идентификатору.
// Проверяет аутентификацию пользователя и гарантирует, что идентификатор пароля правильно анализируется из запроса.
//
//	@Summary		Удалить пароль пользователя
//	@Description	Удаляет пароль пользователя по его идентификатору. Гарантирует, что пользователь аутентифицирован и идентификатор пароля корректен.
//	@Tags			password
//	@Param			id	path	string	true	"Password ID"
//	@Produce		json
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid password ID"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/password/{id} [delete]
func (s *PasswordService) DeleteUserPasswords(response http.ResponseWriter, request *http.Request) {
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
	if err = s.repository.DeleteByUserIDAndID(request.Context(), user.ID, id); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t delete", http.StatusInternalServerError, response)
		return
	}
}

// RegisterRoutes настраивает HTTP-маршруты для PasswordService, применяя предоставленное промежуточное программное обеспечение для обработки паролей.
func (s *PasswordService) RegisterRoutes(middlewares ...func(http.Handler) http.Handler) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/password", s.SavePasswordHandler)
		r.Put("/password", s.UpdatePasswordHandler)
		r.Get("/password", s.GetUserPasswords)
		r.Delete("/password/{id}", s.DeleteUserPasswords)
	}
}
