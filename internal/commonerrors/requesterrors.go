package commonerrors

// RequestError представляет собой ошибку, возникшую во время обработки запроса, с внутренней ошибкой и соответствующим статусом HTTP.
type RequestError struct {
	InternalError error
	HTTPStatus    int
}

func (e *RequestError) Error() string {
	if e.InternalError == nil {
		return "RequestError"
	}
	return e.InternalError.Error()
}

func (e *RequestError) Unwrap() error {
	return e.InternalError
}
