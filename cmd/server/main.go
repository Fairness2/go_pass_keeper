package main

import (
	"log"
	"passkeeper/internal/config"
	"passkeeper/internal/logger"
	"passkeeper/internal/serverapp"
)

func main() {
	log.Println("Start program")
	// Устанавливаем настройки
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger.Log, err = logger.New(cnf.LogLevel)
	if err != nil {
		log.Fatal(err)
	}

	// Показываем конфигурацию сервера
	logger.Log.Infow("Running server with configuration", "config", cnf)

	// стартуем приложение
	if err = serverapp.New(cnf); err != nil {
		logger.Log.Error(err)
	}

	logger.Log.Info("End program")
}
