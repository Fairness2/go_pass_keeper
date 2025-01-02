package content

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi/v5"
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

// textRepository определяет методы управления текстовым содержимым пользователя и связанными с ним комментариями в системе.
type textRepository interface {
	Create(ctx context.Context, content models.TextContent, comment models.Comment) error
	GetByUserIDAndId(ctx context.Context, userID int64, id string) (*models.TextContent, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.TextWithComment, error)
	DeleteByUserIDAndID(ctx context.Context, userID int64, id string) error
}

// TextService предоставляет методы для управления текстами пользователей и соответствующими комментариями в системе.
type TextService struct {
	repository textRepository
}

// NewTextService инициализирует и возвращает новый экземпляр TextService, настроенный с использованием предоставленной базой.
func NewTextService(rep textRepository) *TextService {
	return &TextService{
		repository: rep,
	}
}

// SaveTextHandler обрабатывает HTTP-запросы на сохранение нового текста вместе с дополнительным комментарием для аутентифицированного пользователя.
// Он гарантирует корректность тела запроса, предотвращает предоставление идентификатора и связывает пароль с пользователем.
//
//	@Summary		Сохранить пользовательский текст с комментарием
//	@Description	Создает новую текстовую запись вместе с необязательным комментарием для аутентифицированного пользователя.
//	@Tags			text
//	@Accept			json
//	@Produce		json
//	@Param			data	body	payloads.SaveText	true	"Text data and comment"
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid input or bad request"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/text [post]
func (s *TextService) SaveTextHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	var body payloads.SaveText
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
	pass := models.TextContent{
		UserID:   user.ID,
		TextData: body.TextData,
	}
	comment := models.Comment{
		ContentType: models.TypeText,
		Comment:     body.Comment,
	}
	if err = s.repository.Create(request.Context(), pass, comment); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// UpdateTextHandler обрабатывает HTTP-запросы на обновление существующего текста и связанного с ним комментария для аутентифицированного пользователя.
// Он проверяет тело запроса, проверяет наличие пароля пользователя и обеспечивает правильную обработку обновлений.
//
//	@Summary		Обновить текст пользователя с комментарием
//	@Description	Обновляет существующий текст вместе с комментарием для аутентифицированного пользователя. Гарантирует существование текста и обрабатывает проверку.
//	@Tags			text
//	@Accept			json
//	@Produce		json
//	@Param			data	body	payloads.SaveText	true	"Updated text data and comment"
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid input or bad request"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		404	{object}	payloads.ErrorResponseBody	"Text not found"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/text [put]
func (s *TextService) UpdateTextHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	var body payloads.UpdateText
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
	text := models.TextContent{
		ID:        body.ID,
		UserID:    user.ID,
		TextData:  body.TextData,
		UpdatedAt: time.Now(),
	}
	comment := models.Comment{
		ContentType: models.TypeText,
		Comment:     body.Comment,
		ContentID:   body.ID,
		UpdatedAt:   time.Now(),
	}
	ctx := request.Context()
	// Проверяем есть ли такой пароль у пользователя
	_, err = s.repository.GetByUserIDAndId(ctx, text.UserID, text.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			responsesetters.ProcessResponseWithStatus("Text not found", http.StatusNotFound, response)
			return
		} else {
			responsesetters.SetInternalError(err, response)
			return
		}
	}
	// Обновляем пароль
	if err = s.repository.Create(ctx, text, comment); err != nil {
		responsesetters.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// GetUserTexts обрабатывает запросы HTTP GET для получения всех текстов и комментариев к ним для аутентифицированного пользователя.
// Он извлекает пользователя из контекста запроса, извлекает соответствующие тексты из репозитория и возвращает их в формате JSON.
//
//	@Summary		Получить тексты пользователей с комментариями
//	@Description	Получает список текстов и связанных с ними комментариев для аутентифицированного пользователя.
//	@Tags			text
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		payloads.TextWithComment	"Array of user texts with comments"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authorized"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/text [get]
func (s *TextService) GetUserTexts(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		responsesetters.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	texts, err := s.repository.GetByUserID(request.Context(), user.ID)
	if err != nil {
		responsesetters.SetInternalError(err, response)
		return
	}
	textsData := make([]payloads.TextWithComment, 0, len(texts))
	for _, text := range texts {
		textsData = append(textsData, payloads.TextWithComment{
			Text: payloads.Text{
				ID:       text.ID,
				TextData: text.TextData,
			},
			Comment: text.Comment,
		})
	}
	marshaledBody, err := json.Marshal(textsData)
	if err != nil {
		responsesetters.SetInternalError(err, response)
		return
	}
	if err = responsesetters.SetHTTPResponse(response, http.StatusOK, marshaledBody); err != nil {
		logger.Log.Error(err)
	}
}

// DeleteUserText обрабатывает удаление записи текста пользователя по его идентификатору.
// Проверяет аутентификацию пользователя и гарантирует, что идентификатор пароля правильно анализируется из запроса.
//
//	@Summary		Удалить текст пользователя
//	@Description	Удаляет текст пользователя по его идентификатору. Гарантирует, что пользователь аутентифицирован и идентификатор текста корректен.
//	@Tags			text
//	@Param			id	path	string	true	"Text ID"
//	@Produce		json
//	@Success		200
//	@Failure		400	{object}	payloads.ErrorResponseBody	"Invalid text ID"
//	@Failure		401	{object}	payloads.ErrorResponseBody	"User not authenticated"
//	@Failure		500	{object}	payloads.ErrorResponseBody	"Internal server error"
//	@Router			/api/content/text/{id} [delete]
func (s *TextService) DeleteUserText(response http.ResponseWriter, request *http.Request) {
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

// getIDFromRequest извлекает и проверяет параметр «id» из предоставленного HTTP-запроса.
func getIDFromRequest(request *http.Request) (string, error) {
	id := chi.URLParam(request, "id")
	p := payloads.IDPayload{ID: id}
	ok, err := govalidator.ValidateStruct(p)
	if !ok {
		return "", &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}
	return id, nil
}

// RegisterRoutes настраивает HTTP-маршруты для TextService, применяя предоставленное промежуточное программное обеспечение для обработки текстовых запросов.
func (s *TextService) RegisterRoutes(middlewares ...func(http.Handler) http.Handler) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(middlewares...)
		r.Post("/text", s.SaveTextHandler)
		r.Put("/text", s.UpdateTextHandler)
		r.Get("/text", s.GetUserTexts)
		r.Delete("/text/{id}", s.DeleteUserText)
	}
}
