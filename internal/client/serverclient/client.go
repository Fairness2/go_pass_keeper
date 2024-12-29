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
	client       *resty.Client
	token        string
	refreshToken string
}

// NewClient инициализирует новый RestClient с предоставленным базовым URL-адресом.
func NewClient(baseURL string) (*Client, error) {
	if baseURL == "" {
		return nil, ErrEmptyBaseURL
	}
	c := resty.New()
	c.BaseURL = baseURL
	return &Client{client: c}, nil
}

// SetTokens устанавливает токен доступа и токен обновления для экземпляра клиента.
func (c *Client) SetTokens(token, refreshToken string) {
	c.token = token
	c.refreshToken = refreshToken
}

// GetToken возвращает текущий токен аутентификации, хранящийся в Клиенте.
func (c *Client) GetToken() string {
	return c.token
}

// GetRequest создает и возвращает новый объект resty.Request для создания и отправки HTTP-запросов.
func (c *Client) GetRequest() *resty.Request {
	return c.client.R()
}
