package middleware

import (
	"context"
	"time"

	"ride-sharing-notification/internal/pkg/logging"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor returns a new unary server interceptor that logs gRPC requests
func LoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		// Get or generate tracking IDs from metadata
		md, _ := metadata.FromIncomingContext(ctx)
		requestID := getFirstValue(md, logging.RequestIDKey)
		if requestID == "" {
			requestID = logging.GenerateID()
		}

		correlationID := getFirstValue(md, logging.CorrelationID)
		if correlationID == "" {
			correlationID = logging.GenerateID()
		}

		// Add IDs to context
		ctx = context.WithValue(ctx, logging.RequestIDKey, requestID)
		ctx = context.WithValue(ctx, logging.CorrelationID, correlationID)

		// Get logger with request context
		logger := logging.GetLogger().WithContext(ctx)

		// Log request start
		logger.Info("gRPC request started",
			zap.String("method", info.FullMethod),
			zap.Any("request", logging.MaskSensitiveData(req)),
		)

		// Process request
		resp, err = handler(ctx, req)
		duration := time.Since(start)

		// Prepare standard log fields
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("latency", duration),
		}

		// Add response/error details
		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields,
				zap.String("grpc_code", st.Code().String()),
				zap.String("error", st.Message()),
			)
			logger.Error("gRPC request failed", fields...)
		} else {
			fields = append(fields,
				zap.Any("response", logging.MaskSensitiveData(resp)),
			)
			logger.Info("gRPC request completed", fields...)
		}

		return resp, err
	}
}

// getFirstValue helper to extract first value from metadata
func getFirstValue(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}
