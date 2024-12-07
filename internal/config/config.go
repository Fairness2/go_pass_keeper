package config

import (
	"crypto/rsa"
	_ "embed"
	"time"
)

const (
	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = ":8080"
	// DefaultLogLevel Уровень логирования по умолчанию
	DefaultLogLevel = "info"
	// DefaultDatabaseDSN подключение к базе данных
	DefaultDatabaseDSN = "postgresql://postgres:example@127.0.0.1:/passkeeper"
	// DefaultHashKey ключ шифрования по умолчанию
	DefaultHashKey = "dsbtyrew3!hgsdfvbytrd324"
	// DefaultTokenExpiration Время жизни токена авторизации по умолчанию
	DefaultTokenExpiration = 12 * time.Hour
)

// DefaultPrivateKey Текстовое представление приватного ключа для JWT по умолчанию
//
//go:embed keys/private.pem
var DefaultPrivateKey string

// DefaultPublicKey Текстовое представление публичного ключа для JWT по умолчанию
//
//go:embed keys/public.pem
var DefaultPublicKey string

type JWTKeys struct {
	Public  *rsa.PublicKey
	Private *rsa.PrivateKey
}

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	Address         string        `env:"RUN_ADDRESS"`      // адрес сервера
	LogLevel        string        `env:"LOG_LEVEL"`        // Уровень логирования
	DatabaseDSN     string        `env:"DATABASE_URI"`     // подключение к базе данных
	HashKey         string        `env:"KEY"`              // Ключ для шифрования
	PrivateKey      string        `env:"PKEY"`             // Приватный ключ для JWT
	PublicKey       string        `env:"PUKEY"`            // Публичный ключ для JWT
	JWTKeys         *JWTKeys      `env:"-"`                // Ключи для JWT
	TokenExpiration time.Duration `env:"TOKEN_EXPIRATION"` // Время жизни токена авторизации
}

// NewDefaultConfig инициализация конфигурации приложения
func NewDefaultConfig() *CliConfig {
	return &CliConfig{
		Address:         DefaultServerURL,
		LogLevel:        DefaultLogLevel,
		DatabaseDSN:     DefaultDatabaseDSN,
		HashKey:         DefaultHashKey,
		TokenExpiration: DefaultTokenExpiration,
		PrivateKey:      DefaultPrivateKey,
		PublicKey:       DefaultPublicKey,
	}
}