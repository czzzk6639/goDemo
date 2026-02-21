package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"game-server/internal/config"
	"game-server/internal/repository"
	"game-server/internal/router"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := repository.InitDB(); err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	defer repository.CloseDB()

	r := router.NewRouter()
	mux := r.Setup()

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.HTTPPort)
	log.Printf("HTTP server starting on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
