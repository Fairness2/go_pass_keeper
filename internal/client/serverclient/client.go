package serverclient

import (
	"errors"
	"github.com/go-resty/resty/v2"
)

var ErrEmptyBaseURL = errors.New("empty base url")

type Client struct {
	Client       *resty.Client
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
	return &Client{Client: c}, nil
}

func (c *Client) SetTokens(token, refreshToken string) {
	c.token = token
	c.refreshToken = refreshToken
}
