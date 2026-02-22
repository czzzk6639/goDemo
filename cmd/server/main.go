package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	"game-server/internal/config"
	"game-server/internal/handler"
	"game-server/internal/repository"
	"game-server/internal/router"
	"game-server/pkg/redis"
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

	if err := redis.InitRedis(); err != nil {
		log.Fatalf("Failed to connect redis: %v", err)
	}
	defer redis.CloseRedis()

	tcpHandler := handler.NewTCPHandler()
	go startTCPServer(tcpHandler)

	r := router.NewRouter()
	mux := r.Setup()

	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.HTTPPort)
	log.Printf("HTTP server starting on %s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func startTCPServer(h *handler.TCPHandler) {
	addr := fmt.Sprintf(":%d", config.GlobalConfig.Server.TCPPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to start TCP server: %v", err)
	}
	defer listener.Close()

	log.Printf("TCP server starting on %s", addr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}

		go h.HandleConn(conn)
	}
}
