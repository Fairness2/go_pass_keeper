package config

import (
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
	return cnf, nil
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
	if err := viper.BindEnv("ServerAddress", "SERVER_ADDRESS"); err != nil {
		return err
	}
	if err := viper.BindEnv("LogLevel", "LOG_LEVEL"); err != nil {
		return err
	}
	return nil
}

// bindArg привязывает аргументы командной строки к ключам конфигурации и устанавливает значения по умолчанию с помощью библиотек pflag и viper.
func bindArg() error {
	pflag.StringP("ServerAddress", "a", DefaultServerURL, "server's address and port ")
	pflag.StringP("LogLevel", "l", DefaultLogLevel, "level of logging")
	pflag.Parse()
	return viper.BindPFlags(pflag.CommandLine)
}
