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

// CardService предоставляет методы для управления картами пользователей и соответствующими комментариями в системе.
type CardService struct {
	dbPool repositories.SQLExecutor
}

// NewCardService инициализирует и возвращает новый экземпляр CardService, настроенный с использованием предоставленной базой.
func NewCardService(dbPool repositories.SQLExecutor) *CardService {
	return &CardService{
		dbPool: dbPool,
	}
}

// SaveCardHandler обрабатывает HTTP-запросы на сохранение новогй карты вместе с дополнительным комментарием для аутентифицированного пользователя.
// Он гарантирует корректность тела запроса, предотвращает предоставление идентификатора и связывает карту с пользователем.
func (s *CardService) SaveCardHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	body, err := s.getSaveCardBody(request)
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
	repository := repositories.NewCardRepository(request.Context(), s.dbPool)
	if err = repository.Create(card, comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// getSaveCardBody анализирует и проверяет тело HTTP-запроса для извлечения полезных данных SaveCard или возвращает ошибку.
func (s *CardService) getSaveCardBody(request *http.Request) (*payloads.SaveCard, error) {
	// TODO валидация
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	// Парсим тело в структуру запроса
	var body payloads.SaveCard
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}

	return &body, nil
}

// UpdateCardHandler обрабатывает HTTP-запросы на обновление существующей карты и связанного с ним комментария для аутентифицированного пользователя.
// Он проверяет тело запроса, проверяет наличие пароля пользователя и обеспечивает правильную обработку обновлений.
func (s *CardService) UpdateCardHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	body, err := s.getSaveCardBody(request)
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
	repository := repositories.NewCardRepository(request.Context(), s.dbPool)

	// Проверяем есть ли такой пароль у пользователя
	_, err = repository.GetByUserIDAndId(card.UserID, card.ID)
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
	if err = repository.Create(card, comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

// GetUserCards обрабатывает запросы HTTP GET для получения всех карт и комментариев к ним для аутентифицированного пользователя.
// Он извлекает пользователя из контекста запроса, извлекает соответствующие пароли из репозитория и возвращает их в формате JSON.
func (s *CardService) GetUserCards(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	repository := repositories.NewCardRepository(request.Context(), s.dbPool)
	cards, err := repository.GetByUserID(user.ID)
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
	repository := repositories.NewCardRepository(request.Context(), s.dbPool)
	if err = repository.DeleteByUserIDAndID(user.ID, id); err != nil {
		helpers.ProcessResponseWithStatus("Can`t delete", http.StatusInternalServerError, response)
		return
	}
}
