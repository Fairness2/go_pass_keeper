package content

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"net/http"
	"passkeeper/internal/helpers"
	"passkeeper/internal/logger"
	"passkeeper/internal/models"
	"passkeeper/internal/payloads"
	"passkeeper/internal/repositories"
	"passkeeper/internal/token"
	"strconv"
	"time"
)

// cardRepository — интерфейс, определяющий операции по управлению данными карты и соответствующими комментариями в хранилище.
type cardRepository interface {
	Create(ctx context.Context, content models.CardContent, comment models.Comment) error
	GetByUserIDAndId(ctx context.Context, userID int64, id int64) (*models.CardContent, error)
	GetByUserID(ctx context.Context, userID int64) ([]models.CardWithComment, error)
	DeleteByUserIDAndID(ctx context.Context, userID int64, id int64) error
}

// CardService предоставляет методы для управления картами пользователей и соответствующими комментариями в системе.
type CardService struct {
	repository cardRepository
}

// NewCardService инициализирует и возвращает новый экземпляр CardService, настроенный с использованием предоставленной базой.
func NewCardService(dbPool repositories.SQLExecutor) *CardService {
	return &CardService{
		repository: repositories.NewCardRepository(dbPool),
	}
}

// SaveCardHandler обрабатывает HTTP-запросы на сохранение новогй карты вместе с дополнительным комментарием для аутентифицированного пользователя.
// Он гарантирует корректность тела запроса, предотвращает предоставление идентификатора и связывает карту с пользователем.
//
// @Summary Сохранить карту пользователя с комментарием
// @Description Создает новую запись о карту вместе с необязательным комментарием для аутентифицированного пользователя.
// @Tags card
// @Accept json
// @Produce json
// @Param data body payloads.SaveCard true "Card data and comment"
// @Success 200 {string}
// @Failure 400 {object} payloads.ErrorResponseBody "Invalid input or bad request"
// @Failure 401 {object} payloads.ErrorResponseBody "User not authenticated"
// @Failure 500 {object} payloads.ErrorResponseBody "Internal server error"
// @Router /content/card [post]
func (s *CardService) SaveCardHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	var body payloads.SaveCard
	err := getSaveBody(request, &body)
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
	card := models.CardContent{
		UserID: user.ID,
		Number: body.Number,
		Date:   body.Date,
		Owner:  body.Owner,
		CVV:    body.CVV,
	}
	comment := models.Comment{
		ContentType: models.TypeCard,
		Comment:     body.Comment,
	}
	if err = s.repository.Create(request.Context(), card, comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// UpdateCardHandler обрабатывает HTTP-запросы на обновление существующей карты и связанного с ним комментария для аутентифицированного пользователя.
// Он проверяет тело запроса, проверяет наличие пароля пользователя и обеспечивает правильную обработку обновлений.
//
// @Summary Обновить карту пользователя с комментарием
// @Description Обновляет существующую карту вместе с комментарием для аутентифицированного пользователя. Гарантирует существование карты и обрабатывает проверку.
// @Tags card
// @Accept json
// @Produce json
// @Param data body payloads.SaveCard true "Updated card data and comment"
// @Success 200 {string}
// @Failure 400 {object} payloads.ErrorResponseBody "Invalid input or bad request"
// @Failure 401 {object} payloads.ErrorResponseBody "User not authenticated"
// @Failure 404 {object} payloads.ErrorResponseBody "Card not found"
// @Failure 500 {object} payloads.ErrorResponseBody "Internal server error"
// @Router /content/card [put]
func (s *CardService) UpdateCardHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	var body payloads.SaveCard
	err := getSaveBody(request, &body)
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
	card := models.CardContent{
		ID:        body.ID,
		UserID:    user.ID,
		Number:    body.Number,
		Date:      body.Date,
		Owner:     body.Owner,
		CVV:       body.CVV,
		UpdatedAt: time.Now(),
	}
	comment := models.Comment{
		ContentType: models.TypeCard,
		Comment:     body.Comment,
		ContentID:   body.ID,
		UpdatedAt:   time.Now(),
	}
	ctx := request.Context()
	// Проверяем есть ли такой пароль у пользователя
	_, err = s.repository.GetByUserIDAndId(ctx, card.UserID, card.ID)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			helpers.ProcessResponseWithStatus("Card not found", http.StatusNotFound, response)
			return
		} else {
			helpers.SetInternalError(err, response)
			return
		}
	}
	// Обновляем пароль
	if err = s.repository.Create(ctx, card, comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// GetUserCards обрабатывает запросы HTTP GET для получения всех карт и комментариев к ним для аутентифицированного пользователя.
// Он извлекает пользователя из контекста запроса, извлекает соответствующие пароли из репозитория и возвращает их в формате JSON.
//
// @Summary Получить карты пользователей с комментариями
// @Description Получает список карт и связанных с ними комментариев для аутентифицированного пользователя.
// @Tags card
// @Accept json
// @Produce json
// @Success 200 {array} payloads.CardWithComment "Array of user cards with comments"
// @Failure 401 {object} payloads.ErrorResponseBody "User not authorized"
// @Failure 500 {object} payloads.ErrorResponseBody "Internal server error"
// @Router /content/card [get]
func (s *CardService) GetUserCards(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	cards, err := s.repository.GetByUserID(request.Context(), user.ID)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	cardsData := make([]payloads.CardWithComment, 0, len(cards))
	for _, card := range cards {
		cardsData = append(cardsData, payloads.CardWithComment{
			Card: payloads.Card{
				ID:     card.ID,
				Number: card.Number,
				Date:   card.Date,
				Owner:  card.Owner,
				CVV:    card.CVV,
			},
			Comment: card.Comment,
		})
	}
	marshaledBody, err := json.Marshal(cardsData)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	if err = helpers.SetHTTPResponse(response, http.StatusOK, marshaledBody); err != nil {
		logger.Log.Error(err)
	}
}

// DeleteUserCard обрабатывает удаление записи карты пользователя по его идентификатору.
// Проверяет аутентификацию пользователя и гарантирует, что идентификатор пароля правильно анализируется из запроса.
//
// @Summary Удалить карту пользователя
// @Description Удаляет карту пользователя по его идентификатору. Гарантирует, что пользователь аутентифицирован и идентификатор карты корректен.
// @Tags card
// @Param id path string true "Card ID"
// @Produce json
// @Success 200 {string}
// @Failure 400 {object} payloads.ErrorResponseBody "Invalid card ID"
// @Failure 401 {object} payloads.ErrorResponseBody "User not authenticated"
// @Failure 500 {object} payloads.ErrorResponseBody "Internal server error"
// @Router /content/card/{id} [delete]
func (s *CardService) DeleteUserCard(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	strID := chi.URLParam(request, "id")
	id, err := strconv.ParseInt(strID, 10, 64)
	if err != nil {
		helpers.ProcessResponseWithStatus("Card ID is not correct", http.StatusBadRequest, response)
		return
	}

	if err = s.repository.DeleteByUserIDAndID(request.Context(), user.ID, id); err != nil {
		helpers.ProcessResponseWithStatus("Can`t delete", http.StatusInternalServerError, response)
		return
	}
}
