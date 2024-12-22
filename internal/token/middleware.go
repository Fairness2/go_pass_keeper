package token

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"passkeeper/internal/commonerrors"
	"passkeeper/internal/config"
	"passkeeper/internal/helpers"
	"passkeeper/internal/logger"
	"passkeeper/internal/models"
	"passkeeper/internal/repositories"
	"strconv"
	"strings"
	"time"
)

// Key тип ключей в контексте
type Key string

// UserKey ключ авторизованного пользователя в контексте
var UserKey Key = "user"

var (
	ErrHeaderNotExists = &commonerrors.RequestError{
		InternalError: errors.New("authorization header is not exists"),
		HTTPStatus:    http.StatusUnauthorized,
	}
	ErrTokenDoesntHasUserId = &commonerrors.RequestError{
		InternalError: errors.New("token doesnt has user id"),
		HTTPStatus:    http.StatusUnauthorized,
	}
	ErrUserIdIsIncorrect = &commonerrors.RequestError{
		InternalError: errors.New("user id is incorrect"),
		HTTPStatus:    http.StatusUnauthorized,
	}
	ErrUserNotExists = &commonerrors.RequestError{
		InternalError: errors.New("user does not exist"),
		HTTPStatus:    http.StatusUnauthorized,
	}
)

type UserRepository interface {
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

type Generator interface {
	Parse(tokenString string) (*jwt.Token, error)
}

// Authenticator выполняет аутентификацию и авторизацию пользователей с использованием токенов JWT и пула баз данных SQL
type Authenticator struct {
	jwtKeys         *config.Keys
	tokenExpiration time.Duration
	repository      UserRepository
	generator       Generator
}

// NewAuthenticator создает и возвращает новый экземпляр Authenticator с указанными параметрами подключения к базе данных и токенам JWT.
func NewAuthenticator(dbPool repositories.SQLExecutor, jwtKeys *config.Keys, tokenExpiration time.Duration) *Authenticator {
	return &Authenticator{
		repository: repositories.NewUserRepository(dbPool),
		generator:  NewJWTGenerator(jwtKeys.Private, jwtKeys.Public, tokenExpiration, JWTTypeAccess),
	}
}

// Middleware авторизовываем пользователя по токену и записываем его в контекст
func (a *Authenticator) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Получаем токен
		tknString, err := a.getToken(r)
		if err != nil {
			helpers.ProcessRequestErrorWithBody(err, w)
			return
		}
		// Парсим токен
		tkn, err := a.generator.Parse(tknString)
		if err != nil {
			logger.Log.Info(err)
			helpers.ProcessResponseWithStatus("token is not valid", http.StatusUnauthorized, w)
			return
		}
		// Получаем идентификатор пользователя из токена
		userID, err := a.getUserIdFromToken(tkn)
		if err != nil {
			helpers.ProcessRequestErrorWithBody(err, w)
			return
		}
		// Получаем пользователя
		user, err := a.getUserById(r.Context(), userID)
		if err != nil {
			helpers.ProcessRequestErrorWithBody(err, w)
			return
		}

		newR := r.WithContext(context.WithValue(r.Context(), UserKey, user))
		next.ServeHTTP(w, newR)
	})
}

// getToken извлекает токен носителя из заголовка авторизации HTTP-запроса или возвращает ошибку, если он отсутствует или недействителен.
func (a *Authenticator) getToken(r *http.Request) (string, error) {
	tknString := r.Header.Get("Authorization")
	if tknString == "" || !strings.HasPrefix(tknString, "Bearer ") {
		return "", ErrHeaderNotExists
	}
	tknString = strings.TrimPrefix(tknString, "Bearer ")
	return tknString, nil
}

// getUserIdFromToken извлекает идентификатор пользователя из проанализированного токена JWT или возвращает ошибку, если токен недействителен или имеет неправильный формат.
func (a *Authenticator) getUserIdFromToken(tkn *jwt.Token) (int64, error) {
	idStr, err := tkn.Claims.GetSubject()
	if err != nil {
		return 0, ErrTokenDoesntHasUserId
	}
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, ErrUserIdIsIncorrect
	}
	return userID, nil
}

// getUserById извлекает пользователя по его идентификатору из репозитория базы данных и возвращает пользователя или ошибку, если он не найден.
func (a *Authenticator) getUserById(ctx context.Context, userID int64) (*models.User, error) {
	user, err := a.repository.GetUserByID(ctx, userID)
	if err != nil && errors.Is(err, repositories.ErrNotExist) {
		return nil, ErrUserNotExists
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}
