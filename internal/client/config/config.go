package config

const (
	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = "http://localhost:8080"
	// DefaultLogLevel Уровень логирования по умолчанию
	DefaultLogLevel = "info"
)

// Конфигурация при билде
var (
	BuildDate     string = "N/A"
	BuildCommit   string = "N/A"
	BuildVersion  string = "N/A"
	BuildOS       string = "N/A"
	ServerAddress string = DefaultServerURL // адрес сервера
	LogLevel      string = DefaultLogLevel  // Уровень логирования
)
