package config

const (
	// DefaultServerURL Url сервера получателя метрик по умолчанию
	DefaultServerURL = "http://localhost:8080"
	// DefaultLogLevel Уровень логирования по умолчанию
	DefaultLogLevel = "info"
)

var (
	BuildDate     string = "N/A"            // Представляет метку времени сборки приложения. По умолчанию «N/A», если не установлено во время сборки.
	BuildCommit   string = "N/A"            // Представляет хеш фиксации сборки, по умолчанию имеет значение «N/A», если не указано в процессе сборки.
	BuildVersion  string = "N/A"            // Представляет версию приложения и по умолчанию имеет значение «N/A», если не установлено в процессе сборки.
	BuildOS       string = "N/A"            // Представляет операционную систему, на которой было создано приложение. По умолчанию используется значение «N/A», если оно не установлено во время сборки.
	ServerAddress string = DefaultServerURL // адрес сервера
	LogLevel      string = DefaultLogLevel  // Уровень логирования
)
