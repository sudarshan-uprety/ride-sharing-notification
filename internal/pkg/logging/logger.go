package logging

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	// Context keys for tracking request information
	RequestIDKey  = "x-request-id"
	CorrelationID = "x-correlation-id"
)

var (
	instance *Logger
	once     sync.Once

	// Fields to be masked in logs for security
	sensitiveFields = map[string]struct{}{
		"password":         {},
		"confirm_password": {},
		"access_token":     {},
		"refresh_token":    {},
		"token":            {},
		"pin":              {},
		"credit_card":      {},
		"cvv":              {},
		"authorization":    {},
		"set-cookie":       {},
	}

	// Standard metadata fields that will be included in all logs
	standardFields []zap.Field
)

// Logger wraps zap.Logger
type Logger struct {
	*zap.Logger
}

// LogConfig holds configuration for logger initialization
type LogConfig struct {
	Environment string
	Version     string
	ServiceName string
}

// InitLogger configures the global logger instance with application metadata
func InitLogger(cfg LogConfig) {
	// Set standard fields that will be included in all logs
	standardFields = []zap.Field{
		zap.String("service", cfg.ServiceName),
		zap.String("environment", cfg.Environment),
		zap.String("version", cfg.Version),
	}

	// Create directory if it doesn't exist
	logDir := filepath.Join("log", cfg.Environment, cfg.Version)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic("failed to create log directory: " + err.Error())
	}

	logPath := filepath.Join(logDir, "log.log")

	// Configure the encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Setup log rotation
	fileWriter := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    100, // megabytes
		MaxBackups: 7,
		MaxAge:     30, // days
		Compress:   true,
	})

	// Priority levels for routing logs
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})

	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= zapcore.InfoLevel
	})

	// Multi-sink setup for different log levels
	cores := []zapcore.Core{
		// File output with rotation for all logs
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return true }),
		),
		// Stderr for errors
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stderr),
			highPriority,
		),
		// Stdout for info and below
		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			lowPriority,
		),
	}

	// Create the logger
	zapLogger := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Fields(standardFields...),
	)

	// Set the global instance
	instance = &Logger{zapLogger}
}

// GetLogger returns the singleton logger instance, initializing if necessary
func GetLogger() *Logger {
	once.Do(func() {
		// If the logger hasn't been initialized yet, create with defaults
		if instance == nil {
			defaultCfg := LogConfig{
				Environment: "development",
				Version:     "0.0.0",
				ServiceName: "service",
			}
			InitLogger(defaultCfg)
		}
	})
	return instance
}

// WithContext enhances logger with request context information
func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	fields := []zap.Field{
		zap.String("request_id", getStringFromContext(ctx, RequestIDKey)),
		zap.String("correlation_id", getStringFromContext(ctx, CorrelationID)),
	}

	// Add goroutine ID for debugging concurrent issues
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	fields = append(fields, zap.String("goroutine", strings.Fields(string(buf[:n]))[1]))

	return &Logger{l.Logger.With(fields...)}
}

// Shutdown flushes any buffered log entries
func (l *Logger) Shutdown() error {
	return l.Sync()
}

// Helper to extract string values from context
func getStringFromContext(ctx context.Context, key string) string {
	if val, ok := ctx.Value(key).(string); ok {
		return val
	}
	return ""
}

// MaskSensitiveData recursively masks sensitive fields in any data structure
func MaskSensitiveData(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, val := range v {
			if _, ok := sensitiveFields[strings.ToLower(key)]; ok {
				v[key] = "****"
			} else {
				v[key] = MaskSensitiveData(val)
			}
		}
		return v
	case []interface{}:
		for i, val := range v {
			v[i] = MaskSensitiveData(val)
		}
		return v
	default:
		return data
	}
}
