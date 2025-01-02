package config

import (
	"crypto/rsa"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	ErrNoPrivateKey = errors.New("no private key path specified")
	ErrNoPublicKey  = errors.New("no public key path specified")
)

// NewConfig инициализирует новую консольную конфигурацию, обрабатывает аргументы командной строки
func NewConfig() (*CliConfig, error) {
	// Регистрируем новое хранилище
	cnf := &CliConfig{}
	if err := parseFromViper(cnf); err != nil {
		return nil, err
	}
	// Парсим ключи для JWT токена
	jwtKeys, err := parseKeys(cnf.PrivateJWTKey, cnf.PublicJWTKey)
	if err != nil {
		return nil, err
	}
	cnf.JWTKeys = jwtKeys

	return cnf, nil
}

// parseKeys парсим ключи для JWT токена
func parseKeys(privateKey, publicKey string) (*Keys, error) {
	pkey, pubKey, err := parseRSAKeys(privateKey, publicKey)
	if err != nil {
		return nil, err
	}
	return &Keys{
		Public:  pubKey,
		Private: pkey,
	}, nil
}

// parseRSAKeys получаем ключи для JWT токенов
func parseRSAKeys(privateKey string, publicKey string) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	if privateKey == "" {
		return nil, nil, ErrNoPrivateKey
	}
	if publicKey == "" {
		return nil, nil, ErrNoPublicKey
	}

	pkey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, nil, err
	}

	return pkey, pubKey, nil
}

// parseFromViper анализирует конфигурацию из переменных среды и аргументов командной строки с помощью Viper.
func parseFromViper(cnf *CliConfig) error {
	bindEnv()
	if err := bindArg(); err != nil {
		return err
	}

	return viper.Unmarshal(cnf)
}

// bindEnv привязывает переменные среды к ключам конфигурации Viper, гарантируя, что каждая привязка проверяется на наличие ошибок.
func bindEnv() {
	viper.MustBindEnv("Address", "RUN_ADDRESS")
	viper.MustBindEnv("LogLevel", "LOG_LEVEL")
	viper.MustBindEnv("DatabaseDSN", "DATABASE_URI")
	viper.MustBindEnv("HashKey", "KEY")
	viper.MustBindEnv("PrivateJWTKey", "JPKEY")
	viper.MustBindEnv("PublicJWTKey", "JPUKEY")
	viper.MustBindEnv("TokenExpiration", "TOKEN_EXPIRATION")
	viper.MustBindEnv("UploadPath", "UPLOAD_PATH")
}

// bindArg привязывает аргументы командной строки к ключам конфигурации и устанавливает значения по умолчанию с помощью библиотек pflag и viper.
func bindArg() error {
	pflag.StringP("Address", "a", DefaultServerURL, "address and port to run server")
	pflag.StringP("LogLevel", "l", DefaultLogLevel, "level of logging")
	pflag.StringP("DatabaseDSN", "d", DefaultDatabaseDSN, "database connection")
	pflag.StringP("HashKey", "h", DefaultHashKey, "encrypted key")
	pflag.StringP("PrivateJWTKey", "r", DefaultPrivateJWTKey, "private jwt key")
	pflag.StringP("PublicJWTKey", "b", DefaultPublicJWTKey, "public jwt key")
	pflag.DurationP("TokenExpiration", "t", DefaultTokenExpiration, "token expiration time")
	pflag.StringP("UploadPath", "w", DefaultUploadPath, "file upload path")
	pflag.Parse()
	return viper.BindPFlags(pflag.CommandLine)
}
