package serverclient

import (
	"errors"
	"github.com/go-resty/resty/v2"
)

// ErrEmptyBaseURL возвращается, когда предоставленный базовый URL-адрес для инициализации клиента пуст.
var ErrEmptyBaseURL = errors.New("empty base url")

// Inst — это глобально доступный экземпляр Клиента, обычно используемый для взаимодействия с сервером.
var Inst *Client

// Client представляет клиент Resty со связанными токенами аутентификации для взаимодействия с API.
type Client struct {
	Client       *resty.Client
	Token        string
	RefreshToken string
}

// NewClient инициализирует новый RestClient с предоставленным базовым URL-адресом.
func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		return nil, ErrEmptyBaseURL
	}
	c := resty.New()
	c.BaseURL = baseURL
	return &Client{Client: c}, nil
}

// SetTokens устанавливает токен доступа и токен обновления для экземпляра клиента.
func (c *Client) SetTokens(token, refreshToken string) {
	c.Token = token
	c.RefreshToken = refreshToken
}
