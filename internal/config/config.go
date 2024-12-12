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

var (
	// DefaultPrivateJWTKey Текстовое представление приватного ключа для JWT по умолчанию
	//
	//go:embed keys/jwt/private.pem
	DefaultPrivateJWTKey string
	// DefaultPublicJWTKey Текстовое представление публичного ключа для JWT по умолчанию
	//
	//go:embed keys/jwt/public.pem
	DefaultPublicJWTKey string
)

type Keys struct {
	Public  *rsa.PublicKey
	Private *rsa.PrivateKey
}

var (
	// DefaultPrivateEncryptKey Текстовое представление приватного ключа для шифрования контента по умолчанию
	//
	//go:embed keys/encrypt/private.pem
	DefaultPrivateEncryptKey string
	// DefaultPublicEncryptKey Текстовое представление публичного ключа для шифрования контента по умолчанию
	//
	//go:embed keys/encrypt/public.pem
	DefaultPublicEncryptKey string
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	Address           string        `env:"RUN_ADDRESS"`      // адрес сервера
	LogLevel          string        `env:"LOG_LEVEL"`        // Уровень логирования
	DatabaseDSN       string        `env:"DATABASE_URI"`     // подключение к базе данных
	HashKey           string        `env:"KEY"`              // Ключ для шифрования
	PrivateJWTKey     string        `env:"JPKEY"`            // Приватный ключ для JWT
	PublicJWTKey      string        `env:"JPUKEY"`           // Публичный ключ для JWT
	PrivateEncryptKey string        `env:"EPKEY"`            // Приватный ключ для шифрования
	PublicEncryptKey  string        `env:"EPUKEY"`           // Публичный ключ для шифрования
	JWTKeys           *Keys         `env:"-"`                // Ключи для JWT
	EncryptKeys       *Keys         `env:"-"`                // Ключи для шифрования
	TokenExpiration   time.Duration `env:"TOKEN_EXPIRATION"` // Время жизни токена авторизации
}
