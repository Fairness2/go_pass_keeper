package content

import (
	"encoding/json"
	"io"
	"net/http"
	"passkeeper/internal/commonerrors"
	"passkeeper/internal/config"
	"passkeeper/internal/helpers"
	"passkeeper/internal/logger"
	"passkeeper/internal/models"
	"passkeeper/internal/payloads"
	"passkeeper/internal/repositories"
	"passkeeper/internal/token"
)

type PasswordService struct {
	dbPool repositories.SQLExecutor
	keys   *config.Keys
}

func NewPasswordService(dbPool repositories.SQLExecutor, keys *config.Keys) *PasswordService {
	return &PasswordService{
		dbPool: dbPool,
		keys:   keys,
	}
}

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
	//encryper := encrypt.NewEncrypter(s.keys.Public)
	pass := models.PasswordContent{
		UserID:   user.ID,
		Domen:    body.Domen,
		Username: body.Username,
		Password: body.Password,
	}
	//if err = pass.EncryptPrivateFields(encryper); err != nil {
	//	helpers.ProcessResponseWithStatus("Can`t encrypt", http.StatusInternalServerError, response)
	//	return
	//}
	comment := models.Comment{
		ContentType: models.TypePassword,
		Comment:     body.Comment,
	}
	repository := repositories.NewPasswordRepository(request.Context(), s.dbPool)
	if err = repository.Create(&pass, &comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

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
	//encryper := encrypt.NewEncrypter(s.keys.Public)
	pass := models.PasswordContent{
		ID:       body.ID,
		UserID:   user.ID,
		Domen:    body.Domen,
		Username: body.Username,
		Password: body.Password,
	}
	/*if err = pass.EncryptPrivateFields(encryper); err != nil {
		helpers.ProcessResponseWithStatus("Can`t encrypt", http.StatusInternalServerError, response)
		return
	}*/
	comment := models.Comment{
		ContentType: models.TypePassword,
		Comment:     body.Comment,
		ContentID:   body.ID,
	}
	repository := repositories.NewPasswordRepository(request.Context(), s.dbPool)
	if err = repository.Create(&pass, &comment); err != nil {
		helpers.ProcessResponseWithStatus("Can`t save", http.StatusInternalServerError, response)
	}
}

func (s *PasswordService) GetUserPasswords(response http.ResponseWriter, request *http.Request) {
	// Берём авторизованного пользователя
	user, ok := request.Context().Value(token.UserKey).(*models.User)
	if !ok {
		helpers.ProcessResponseWithStatus("User not found", http.StatusUnauthorized, response)
		return
	}
	repository := repositories.NewPasswordRepository(request.Context(), s.dbPool)
	passwords, err := repository.GetPasswordsByUserID(user.ID)
	if err != nil {
		helpers.SetInternalError(err, response)
		return
	}
	//decryptor := encrypt.NewDecrypter(s.keys.Private)
	passwordsData := make([]payloads.PasswordWithComment, 0, len(passwords))
	for _, password := range passwords {
		//if err = password.DecryptPrivateFields(decryptor); err != nil {
		//	helpers.SetInternalError(err, response)
		//	return
		//}
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
