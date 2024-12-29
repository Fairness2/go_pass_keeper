package user

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/asaskevich/govalidator"
	"io"
	"net/http"
	"passkeeper/internal/commonerrors"
	"passkeeper/internal/config"
	"passkeeper/internal/helpers"
	"passkeeper/internal/models"
	"passkeeper/internal/payloads"
	"passkeeper/internal/repositories"
	"passkeeper/internal/token"
	"time"
)

var (
	ErrUserAlreadyExists      = errors.New("user already exists")
	ErrLoginPasswordIncorrect = errors.New("login or password is incorrect")
	ErrBadRequest             = errors.New("bad request")
)

type repository interface {
	UserExists(ctx context.Context, login string) error
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
}

// Handlers для обработки запросов, связанных с регистрацией и аутентификацией пользователей.
type Handlers struct {
	repository             repository
	jwtKeys                *config.Keys
	tokenExpiration        time.Duration
	refreshTokenExpiration time.Duration
	hashKey                string
}

type HandlerConfig struct {
	DBPool                 repositories.SQLExecutor
	JwtKeys                *config.Keys
	TokenExpiration        time.Duration
	RefreshTokenExpiration time.Duration
	HashKey                string
}

// NewHandlers инициализирует и возвращает новый экземпляр Handlers,
// настроенный с указанным подключением к базе данных, ключами JWT, сроком действия токена и хэш-ключом.
func NewHandlers(conf HandlerConfig) *Handlers {
	return &Handlers{
		repository:             repositories.NewUserRepository(conf.DBPool),
		jwtKeys:                conf.JwtKeys,
		tokenExpiration:        conf.TokenExpiration,
		refreshTokenExpiration: conf.RefreshTokenExpiration,
		hashKey:                conf.HashKey,
	}
}

// RegistrationHandler обрабатывает регистрацию новых пользователей, включая проверку, создание и генерацию токенов.
//
//	@Summary		Регистрация нового пользователя
//	@Description	обрабатывает регистрацию новых пользователей, включая проверку, создание и генерацию токенов.
//	@Tags			Пользователь
//	@Accept			json
//	@Produce		json
//	@Param			register	body		payloads.Register	true	"Register Payload"
//	@Success		200			{object}	payloads.Authorization
//	@Failure		400			{object}	payloads.ErrorResponseBody
//	@Failure		409			{object}	payloads.ErrorResponseBody
//	@Failure		500			{object}	payloads.ErrorResponseBody
//	@Router			/api/user/register [post]
func (l *Handlers) RegistrationHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	body, err := l.getBody(request)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}

	ctx := request.Context()
	// Проверим есть ли пользователь с таким логином
	err = l.userExists(ctx, body.Login)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}

	// Создаём и регистрируем пользователя
	user, err := l.createAndSaveUser(ctx, body)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}

	// Создаём токен для пользователя
	payload, err := l.createTokens(user)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}

	l.setResponse(payload, response)
}

// setResponse — это метод, который записывает полезную нагрузку авторизации в ответ HTTP с соответствующими заголовками и статусом.
func (l *Handlers) setResponse(payload payloads.Authorization, response http.ResponseWriter) {
	responseBody, err := json.Marshal(payload)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	response.Header().Set("Authorization", "Bearer "+payload.Token)
	if rErr := helpers.SetHTTPResponse(response, http.StatusOK, responseBody); rErr != nil {
		helpers.SetInternalError(err, response)
	}
}

// createTokens генерирует токены доступа и обновления для данного пользователя и возвращает их в полезных данных авторизации.
func (l *Handlers) createTokens(user *models.User) (payloads.Authorization, error) {
	// Создаём токен для пользователя
	tkn, err := l.createJWTToken(user, token.JWTTypeAccess, l.tokenExpiration)
	if err != nil {
		return payloads.Authorization{}, err
	}
	// Создаём рефреш токен для пользователя
	refreshTkn, err := l.createJWTToken(user, token.JWTTypeRefresh, l.refreshTokenExpiration)
	if err != nil {
		return payloads.Authorization{}, err
	}
	return payloads.Authorization{
		Token:   tkn,
		Refresh: refreshTkn,
	}, nil
}

// getBody получаем тело для регистрации
func (l *Handlers) getBody(request *http.Request) (*payloads.Register, error) {
	// Читаем тело запроса
	rawBody, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	// Парсим тело в структуру запроса
	var body payloads.Register
	err = json.Unmarshal(rawBody, &body)
	if err != nil {
		return nil, &commonerrors.RequestError{InternalError: err, HTTPStatus: http.StatusBadRequest}
	}

	result, err := govalidator.ValidateStruct(body)
	if err != nil {
		return nil, err
	}

	if !result {
		return nil, &commonerrors.RequestError{InternalError: ErrBadRequest, HTTPStatus: http.StatusBadRequest}
	}

	return &body, nil
}

// createUser создаём нового пользователя
func (l *Handlers) createUser(body *payloads.Register) (*models.User, error) {
	user := &models.User{
		Login:    body.Login,
		Password: body.Password,
	}
	err := user.GeneratePasswordHash(l.hashKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// createAndSaveUser создаём и сохраняем нового пользователя
func (l *Handlers) createAndSaveUser(ctx context.Context, body *payloads.Register) (*models.User, error) {
	user, err := l.createUser(body)
	if err != nil {
		return nil, err
	}
	if err = l.repository.CreateUser(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// createJWTToken создаём JWT токен
func (l *Handlers) createJWTToken(user *models.User, tokenType token.JWTType, expiration time.Duration) (string, error) {
	generator := token.NewJWTGenerator(l.jwtKeys.Private, l.jwtKeys.Public, expiration, tokenType)
	return generator.Generate(user)
}

// LoginHandler обрабатывает вход пользователя в систему, проверяя учетные данные и генерируя токен авторизации.
//
//	@Summary		Вход пользователя в систему
//	@Description	обрабатывает вход пользователя в систему, проверяя учетные данные и генерируя токен авторизации.
//	@Tags			Пользователь
//	@Accept			json
//	@Produce		json
//	@Param			login	body		payloads.Register	true	"Login Payload"
//	@Success		200		{object}	payloads.Authorization
//	@Failure		400		{object}	payloads.ErrorResponseBody
//	@Failure		401		{object}	payloads.ErrorResponseBody
//	@Failure		500		{object}	payloads.ErrorResponseBody
//	@Router			/api/user/login [post]
func (l *Handlers) LoginHandler(response http.ResponseWriter, request *http.Request) {
	// Читаем тело запроса
	body, err := l.getBody(request)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}
	requestedUser, err := l.createUser(body)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}

	dbUser, err := l.getUserByLogin(request.Context(), requestedUser.Login)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}

	err = l.checkPassword(dbUser, requestedUser.PasswordHash)
	if err != nil {
		helpers.ProcessRequestErrorWithBody(err, response)
		return
	}

	// Создаём токен для пользователя
	payload, err := l.createTokens(dbUser)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	l.setResponse(payload, response)
}

// checkPassword проверяет, соответствует ли предоставленный хэш пароля сохраненному хэшу пользователя.
// Возвращает ошибку, если проверка не удалась или произошла внутренняя ошибка.
func (l *Handlers) checkPassword(user *models.User, passwordHash string) error {
	ok, err := user.CheckPasswordHash(passwordHash)
	if err != nil {
		return &commonerrors.RequestError{
			InternalError: err,
			HTTPStatus:    http.StatusInternalServerError,
		}
	}
	if !ok {
		return &commonerrors.RequestError{
			InternalError: ErrLoginPasswordIncorrect,
			HTTPStatus:    http.StatusUnauthorized,
		}
	}
	return nil
}

// getUserByLogin извлекает пользователя из репозитория, используя предоставленные учетные данные для входа.
// Возвращает ошибку, если пользователь не существует или произошла внутренняя ошибка сервера.
func (l *Handlers) getUserByLogin(ctx context.Context, login string) (*models.User, error) {
	dbUser, err := l.repository.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			return nil, &commonerrors.RequestError{
				InternalError: ErrLoginPasswordIncorrect,
				HTTPStatus:    http.StatusUnauthorized,
			}
		}
		return nil, &commonerrors.RequestError{
			InternalError: err,
			HTTPStatus:    http.StatusInternalServerError,
		}
	}
	return dbUser, nil
}

// userExists проверяет, существует ли в репозитории пользователь с указанным логином.
// Возвращает ошибку конфликта, если пользователь существует, или внутреннюю ошибку сервера в случае других проблем.
func (l *Handlers) userExists(ctx context.Context, login string) error {
	// Проверим есть ли пользователь с таким логином
	err := l.repository.UserExists(ctx, login)
	if err != nil {
		if errors.Is(err, repositories.ErrNotExist) {
			return nil
		}
		return &commonerrors.RequestError{
			InternalError: err,
			HTTPStatus:    http.StatusInternalServerError,
		}
	}
	return &commonerrors.RequestError{
		InternalError: ErrUserAlreadyExists,
		HTTPStatus:    http.StatusConflict,
	}
}
