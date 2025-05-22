package main

import (
	"log"
	"ride-sharing-notification/config"
	"ride-sharing-notification/internal/pkg/logging"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logging.InitLogger(logging.LogConfig{
		Environment: cfg.Log.Environment,
		Version:     cfg.Log.Version,
		ServiceName: cfg.Log.ServiceName,
	})

	// Initialize services
	// emailSvc := email.NewService(cfg.Email)
	// pushSvc := firebase.NewService(cfg.Firebase)

	// Graceful shutdown
	// quit := make(chan os.Signal, 1)
	// signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

}
