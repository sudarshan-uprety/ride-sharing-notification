package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server struct {
		Port        string
		Environment string
	}
	JWT struct {
		AccessSecret  string
		RefreshSecret string
	}
	Log struct {
		Environment string
		Version     string
		ServiceName string
	}
	Email struct {
		Enabled   bool
		FromEmail string
		Username  string
		Password  string
		SMTPHost  string
		SMTPPort  string
		Timeout   time.Duration
	}
	Kafka struct {
		Brokers  []string
		Topic    string
		Balancer string
		GroupId  string
	}
	GRPC struct {
		Port string
	}
}

func Load() (*Config, error) {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found - using system environment variables")
	}

	cfg := &Config{}

	// Server configuration
	cfg.Server.Port = getEnv("SERVER_PORT", "8080")
	cfg.Server.Environment = getEnv("ENVIRONMENT", "dev")

	// Log configuration
	cfg.Log.Environment = getEnv("ENVIRONMENT", "dev")
	cfg.Log.Version = getEnv("VERSION", "1.0.0")
	cfg.Log.ServiceName = getEnv("SERVICE_NAME", "notification-service")

	// Email configuration (Zoho Mail)
	cfg.Email.Enabled = getEnvAsBool("EMAIL_ENABLED", true)
	cfg.Email.FromEmail = getEnv("EMAIL_FROM", "no-reply@yourdomain.com")
	cfg.Email.Username = getEnv("EMAIL_USERNAME", "your-email@yourdomain.com")
	cfg.Email.Password = getEnv("EMAIL_PASSWORD", "your-zoho-password")
	cfg.Email.SMTPHost = getEnv("EMAIL_SMTP_HOST", "smtp.zoho.com")
	cfg.Email.SMTPPort = getEnv("EMAIL_SMTP_PORT", "587")
	cfg.Email.Timeout = getEnvAsDuration("EMAIL_TIMEOUT", 10*time.Second)

	// Kafka configuration
	cfg.Kafka.Brokers = []string{getEnv("KAFKA_BROKER", "localhost:9092")}
	cfg.Kafka.Topic = getEnv("KAFKA_TOPIC", "default-topic")
	cfg.Kafka.Balancer = getEnv("KAFKA_BALANCER", "least-bytes")
	cfg.Kafka.GroupId = getEnv("KAFKA_GROUP_ID", "user-events-reader")

	// gRPC configuration
	cfg.GRPC.Port = getEnv("GRPC_PORT", "50051")

	return cfg, nil
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if dur, err := time.ParseDuration(value); err == nil {
			return dur
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string, sep string) []string {
	if value, exists := os.LookupEnv(key); exists {
		return strings.Split(value, sep)
	}
	return defaultValue
}
