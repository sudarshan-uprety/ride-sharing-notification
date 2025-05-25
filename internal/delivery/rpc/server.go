package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"
	"ride-sharing-notification/internal/pkg/logging"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	UnimplementedNotificationServiceServer
	emailService EmailServiceClient
	pushService  PushServiceClient
	grpcServer   *grpc.Server
}

func NewServer(
	emailService EmailServiceClient,
) *Server {
	return &Server{
		emailService: emailService,
	}
}

func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer(
		grpc.ConnectionTimeout(5*time.Second),
		grpc.ChainUnaryInterceptor(
			s.contextInterceptor,
			s.loggingInterceptor,
			s.errorHandlingInterceptor,
		),
	)
	RegisterNotificationServiceServer(s.grpcServer, s)

	logging.GetLogger().Info("gRPC server started", zap.String("port", port))
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop(ctx context.Context) {
	if s.grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			s.grpcServer.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
			s.grpcServer.Stop()
			logging.GetLogger().Warn("gRPC server forced to stop")
		case <-stopped:
			logging.GetLogger().Info("gRPC server stopped gracefully")
		}
	}
}

func (s *Server) SendEmail(ctx context.Context, req *EmailRequest) (*NotificationResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	_, err := s.emailService.SendEmail(ctx, req)
	if err != nil {
		return nil, err
	}

	return &NotificationResponse{
		Success:        true,
		Message:        "Email sent successfully",
		NotificationId: generateID(),
	}, nil
}

func (s *Server) SendPush(ctx context.Context, req *PushRequest) (*NotificationResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	_, err := s.pushService.SendPush(ctx, req)
	if err != nil {
		return nil, err
	}

	return &NotificationResponse{
		Success:        true,
		Message:        "Push notification sent successfully",
		NotificationId: generateID(),
	}, nil
}

// contextInterceptor extracts request metadata and adds to context
func (s *Server) contextInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	// Extract headers from incoming metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	// Get or generate request ID
	requestID := getFirstValue(md, "x-request-id")
	if requestID == "" {
		requestID = generateID()
	}

	// Get or generate correlation ID
	correlationID := getFirstValue(md, "x-correlation-id")
	if correlationID == "" {
		correlationID = generateID()
	}

	// Add to context
	ctx = context.WithValue(ctx, logging.RequestIDKey, requestID)
	ctx = context.WithValue(ctx, logging.CorrelationID, correlationID)

	return handler(ctx, req)
}

// loggingInterceptor handles all request logging
func (s *Server) loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()
	logger := logging.GetLogger().WithContext(ctx)

	// Log request start
	logger.Info("gRPC request started",
		zap.String("method", info.FullMethod),
		zap.Any("request", logging.MaskSensitiveData(req)),
	)

	// Call handler
	resp, err = handler(ctx, req)
	duration := time.Since(start)

	// Prepare log fields
	fields := []zap.Field{
		zap.String("method", info.FullMethod),
		zap.Duration("duration", duration),
	}

	// Add response/error details
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			fields = append(fields,
				zap.String("grpc_code", st.Code().String()),
				zap.String("error", st.Message()),
			)
		} else {
			fields = append(fields, zap.Error(err))
		}
		logger.Error("gRPC request failed", fields...)
	} else {
		fields = append(fields,
			zap.Any("response", logging.MaskSensitiveData(resp)),
		)
		logger.Info("gRPC request completed", fields...)
	}

	return resp, err
}

// errorHandlingInterceptor converts errors to gRPC status errors
func (s *Server) errorHandlingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		// Convert errors to gRPC status errors
		var code codes.Code
		switch {
		case errors.Is(err, context.DeadlineExceeded):
			code = codes.DeadlineExceeded
		case errors.Is(err, context.Canceled):
			code = codes.Canceled
		default:
			code = codes.Internal
		}

		return nil, status.Errorf(code, "request failed: %v", err)
	}
	return resp, nil
}

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func getFirstValue(md metadata.MD, key string) string {
	vals := md.Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}
