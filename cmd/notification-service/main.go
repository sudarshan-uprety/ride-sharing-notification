package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"ride-sharing-notification/config"
	"ride-sharing-notification/internal/delivery/kafka"
	"ride-sharing-notification/internal/delivery/rpc"
	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/pkg/logging"
	"syscall"
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
	// Create gRPC server
	grpcServer := rpc.NewGRPCServer(emailSvc)

	// Start server in a goroutine
	go func() {
		if err := grpcServer.Start(cfg.GRPC.Port); err != nil {
		}
	}()
	// Start Kafka consumer
	ctx, cancel := context.WithCancel(context.Background())
	kafkaHandler := kafka.NewMessageHandler(emailSvc)
	consumer := kafka.NewConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupId, kafkaHandler)

	go consumer.Start(ctx)
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal
	<-quit

	// Create shutdown context with timeout
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Gracefully stop the server
	grpcServer.Stop(ctx)
}
