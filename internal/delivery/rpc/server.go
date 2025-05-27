package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"ride-sharing-notification/internal/pkg/logging"
	"ride-sharing-notification/internal/pkg/response"
	"ride-sharing-notification/internal/proto/notification"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	anypb "google.golang.org/protobuf/types/known/anypb"
)

type Server struct {
	notification.UnimplementedNotificationServiceServer
	emailService EmailServiceClient
	pushService  PushServiceClient
	grpcServer   *grpc.Server
}

func NewServer(
	emailService EmailServiceClient,
	pushService PushServiceClient,
) *Server {
	return &Server{
		emailService: emailService,
		pushService:  pushService,
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
	notification.RegisterNotificationServiceServer(s.grpcServer, s)

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

func (s *Server) SendEmail(ctx context.Context, req *notification.EmailRequest) (*notification.StandardResponse, error) {
	if req == nil {
		return response.New().
			Error(codes.InvalidArgument).
			WithMessage("request cannot be nil").
			SimpleError("INVALID_REQUEST"), nil
	}

	// Call the email service
	_, err := s.emailService.SendEmail(ctx, req)
	if err != nil {
		// Convert the error to a standard error response
		st, ok := status.FromError(err)
		if !ok {
			st = status.New(codes.Internal, err.Error())
		}

		return response.New().
			Error(st.Code()).
			WithMessage(st.Message()).
			WithError("EMAIL_SEND_FAILED", nil), nil
	}

	// Create success response with the notification data
	notificationResp := &StandardResponse{
		Success: true,
		Message: "Email sent successfully",
	}

	// Convert to Any type
	anyPayload, err := anypb.New(notificationResp)
	if err != nil {
		return response.New().
			Error(codes.Internal).
			WithMessage("failed to create response").
			WithError("INTERNAL_ERROR", nil), nil
	}

	return response.New().
		Success().
		WithMessage("Email sent successfully").
		WithData(anyPayload, nil)
}

func (s *Server) SendPush(ctx context.Context, req *notification.PushRequest) (*notification.StandardResponse, error) {
	if req == nil {
		return response.New().
			Error(codes.InvalidArgument).
			WithMessage("request cannot be nil").
			SimpleError("INVALID_REQUEST"), nil
	}

	// Call the push service
	_, err := s.pushService.SendPush(ctx, req)
	if err != nil {
		// Convert the error to a standard error response
		st, ok := status.FromError(err)
		if !ok {
			st = status.New(codes.Internal, err.Error())
		}

		return response.New().
			Error(st.Code()).
			WithMessage(st.Message()).
			WithError("PUSH_SEND_FAILED", nil), nil
	}

	// Create success response with the notification data
	notificationResp := &StandardResponse{
		Success: true,
		Message: "Push notification sent successfully",
	}

	// Convert to Any type
	anyPayload, err := anypb.New(notificationResp)
	if err != nil {
		return response.New().
			Error(codes.Internal).
			WithMessage("failed to create response").
			WithError("INTERNAL_ERROR", nil), nil
	}

	return response.New().
		Success().
		WithMessage("Push notification sent successfully").
		WithData(anyPayload, nil)
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
		// If the error is already a gRPC status error, return it directly
		if _, ok := status.FromError(err); ok {
			return nil, err
		}

		// Convert other errors to standard response
		builder := response.New().
			Error(codes.Internal).
			WithMessage(err.Error())

		return builder.WithError("INTERNAL_ERROR", nil), nil
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
