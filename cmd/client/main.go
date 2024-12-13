package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"log"
	"os"
	"passkeeper/internal/client/config"
	"passkeeper/internal/client/models/login"
	"passkeeper/internal/client/serverclient"
	"passkeeper/internal/client/service"
)

func main() {
	// Устанавливаем настройки TODO ставить настройки во время билда?
	cnf, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	serverclient.Inst, err = serverclient.NewClient(cnf.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}
	lgnService := service.NewLoginService(serverclient.Inst)
	initialModel := login.InitialModel(lgnService)
	if _, err := tea.NewProgram(initialModel, tea.WithAltScreen()).Run(); err != nil {
		fmt.Printf("could not start program: %s\n", err)
		os.Exit(1)
	}
}
