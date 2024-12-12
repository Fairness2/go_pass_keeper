package config

import (
	_ "embed"
)

const (
	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = "http://localhost:8080"
	// DefaultLogLevel Уровень логирования по умолчанию
	DefaultLogLevel = "info"
)

// CliConfig конфигурация сервера из командной строки
type CliConfig struct {
	ServerAddress string `env:"SERVER_ADDRESS"` // адрес сервера
	LogLevel      string `env:"LOG_LEVEL"`      // Уровень логирования
}
