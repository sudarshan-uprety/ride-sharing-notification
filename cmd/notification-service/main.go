package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ride-sharing-notification/configs"
	"ride-sharing-notification/internal/delivery/kafka"
	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/pkg/firebase"
	"ride-sharing-notification/internal/pkg/kafka"
	"ride-sharing-notification/internal/pkg/logger"

	"github.com/yourorg/ride-sharing-notification/internal/delivery/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	// Load config
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Initialize logger
	zapLogger, err := logger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	defer zapLogger.Sync()

	// Initialize services
	emailSvc := email.NewService(cfg.Email)
	pushSvc := firebase.NewService(cfg.Firebase)

	// Initialize Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.Brokers, zapLogger)
	if err != nil {
		zapLogger.Fatal("failed to create kafka producer", zap.Error(err))
	}
	defer kafkaProducer.Close()

	// Initialize gRPC server
	grpcServer := grpc.NewServer(kafkaProducer, zapLogger, emailSvc, pushSvc)

	// Initialize Kafka consumer
	kafkaConsumer := kafka.NewConsumer(
		cfg.Kafka.Brokers,
		"notification-fallback",
		"notification-service-group",
		zapLogger,
		grpcServer,
	)

	go func() {
		if err := grpcServer.Start(cfg.GRPCPort); err != nil {
			zapLogger.Fatal("failed to start gRPC server", zap.Error(err))
		}
	}()

	go func() {
		if err := kafkaConsumer.Run(context.Background()); err != nil {
			zapLogger.Fatal("failed to start Kafka consumer", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	zapLogger.Info("server exited properly")
}
