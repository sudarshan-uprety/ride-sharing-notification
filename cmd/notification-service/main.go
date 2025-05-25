package main

import (
	"log"
	"ride-sharing-notification/config"
	"ride-sharing-notification/internal/delivery/rpc"
	"ride-sharing-notification/internal/pkg/email"
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
	emailSvc := email.NewService(cfg)
	// pushSvc, err := firebase.NewService(cfg)

	grpcServer := rpc.NewServer(emailSvc, rpc.EmailServiceClient)
}
