package serverclient

import (
	"errors"
	"github.com/go-resty/resty/v2"
)

var ErrEmptyBaseURL = errors.New("empty base url")

var Inst *Client

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

func (c *Client) SetTokens(token, refreshToken string) {
	c.Token = token
	c.RefreshToken = refreshToken
}
