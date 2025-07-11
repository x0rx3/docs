package main

import (
	"docs/internal/config"
	"docs/internal/logging"
	"docs/pkg/database"
	"docs/pkg/http"
	"docs/pkg/service"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func main() {
	configPath := flag.String("config", "config/config.yaml", "path to config.yaml file")
	flag.Parse()

	config, err := config.NewConfig(*configPath)
	if err != nil {
		fmt.Printf("failed read config: %s", err.Error())
		os.Exit(1)
	}

	log := logging.InitLogger(config.LogLevel)

	repo, err := database.NewPostresRepository(log, config.DSN)
	if err != nil {
		log.Error("failed connect to db", zap.Error(err))
		os.Exit(1)
	}

	serviceCollector := service.NewServiceCollector(log, config.UploadPath, config.AdminToken, repo)

	if err := http.NewServer(log, serviceCollector).Start(config.Addresss, config.Port); err != nil {
		log.Error("failed start listening", zap.Error(err))
		os.Exit(1)
	}

}
