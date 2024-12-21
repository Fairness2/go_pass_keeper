package config

import (
	"crypto/rsa"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	// Парсим ключи для шифрования
	enKeys, err := parseKeys(cnf.PrivateEncryptKey, cnf.PublicEncryptKey)
	if err != nil {
		return nil, err
	}
	cnf.EncryptKeys = enKeys

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
		return nil, nil, errors.New("no private key path specified")
	}
	if publicKey == "" {
		return nil, nil, errors.New("no public key path specified")
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
	if err := bindEnv(); err != nil {
		return err
	}
	if err := bindArg(); err != nil {
		return err
	}

	return viper.Unmarshal(cnf)
}

// bindEnv привязывает переменные среды к ключам конфигурации Viper, гарантируя, что каждая привязка проверяется на наличие ошибок.
func bindEnv() error {
	if err := viper.BindEnv("Address", "RUN_ADDRESS"); err != nil {
		return err
	}
	if err := viper.BindEnv("LogLevel", "LOG_LEVEL"); err != nil {
		return err
	}
	if err := viper.BindEnv("DatabaseDSN", "DATABASE_URI"); err != nil {
		return err
	}
	if err := viper.BindEnv("HashKey", "KEY"); err != nil {
		return err
	}
	if err := viper.BindEnv("PrivateJWTKey", "JPKEY"); err != nil {
		return err
	}
	if err := viper.BindEnv("PublicJWTKey", "JPUKEY"); err != nil {
		return err
	}
	if err := viper.BindEnv("PrivateEncryptKey", "EPKEY"); err != nil {
		return err
	}
	if err := viper.BindEnv("PublicEncryptKey", "EPUKEY"); err != nil {
		return err
	}
	if err := viper.BindEnv("TokenExpiration", "TOKEN_EXPIRATION"); err != nil {
		return err
	}
	if err := viper.BindEnv("UploadPath", "UPLOAD_PATH"); err != nil {
		return err
	}
	return nil
}

// bindArg привязывает аргументы командной строки к ключам конфигурации и устанавливает значения по умолчанию с помощью библиотек pflag и viper.
func bindArg() error {
	pflag.StringP("Address", "a", DefaultServerURL, "address and port to run server")
	pflag.StringP("LogLevel", "l", DefaultLogLevel, "level of logging")
	pflag.StringP("DatabaseDSN", "d", DefaultDatabaseDSN, "database connection")
	pflag.StringP("HashKey", "h", DefaultHashKey, "encrypted key")
	pflag.StringP("PrivateJWTKey", "r", DefaultPrivateJWTKey, "private jwt key")
	pflag.StringP("PublicJWTKey", "b", DefaultPublicJWTKey, "public jwt key")
	pflag.StringP("PrivateEncryptKey", "p", DefaultPrivateEncryptKey, "private encrypt key")
	pflag.StringP("PublicEncryptKey", "u", DefaultPublicEncryptKey, "public encrypt key")
	pflag.DurationP("TokenExpiration", "t", DefaultTokenExpiration, "token expiration time")
	pflag.StringP("UploadPath", "w", DefaultUploadPath, "file upload path")
	pflag.Parse()
	return viper.BindPFlags(pflag.CommandLine)
}
