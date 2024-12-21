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
	var err error
	//bildStr := fmt.Sprintf("BuildDate: %s\nBuildCommit: %s\nBuildVersion: %s\nBuildOS: %s\nServerAddress: %s\nLogLevel: %s\n", config.BuildDate, config.BuildCommit, config.BuildVersion, config.BuildOS, config.ServerAddress, config.LogLevel)
	//fmt.Println(bildStr)
	serverclient.Inst, err = serverclient.NewClient(config.ServerAddress)
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
