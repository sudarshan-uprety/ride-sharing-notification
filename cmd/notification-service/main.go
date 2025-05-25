package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"ride-sharing-notification/config"
	"ride-sharing-notification/internal/delivery/rpc"
	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/pkg/logging"
	"syscall"
	"time"
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

	grpcServer := rpc.NewServer(emailSvc)
	go func() {
		if err := grpcServer.Start(cfg.Server.Port); err != nil {
			fmt.Println("error occured", err.Error())
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	grpcServer.Stop(ctx)

}
