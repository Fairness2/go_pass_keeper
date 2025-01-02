package payloads

// ErrorResponseBody тело ответа с ошибкой
type ErrorResponseBody struct {
	Status  int    `json:"status"` // Успешный или не успешный результат
	Message string `json:"message,omitempty"`
}

type IDPayload struct {
	ID string `json:"id" valid:"required,uuidv4"`
}
