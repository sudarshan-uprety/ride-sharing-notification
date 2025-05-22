package rpc

import (
	"context"
	"math/rand"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	UnimplementedNotificationServiceServer
	logger       *zap.Logger
	emailService EmailServiceClient
	pushService  PushServiceClient
	grpcServer   *grpc.Server
}

func NewServer(
	logger *zap.Logger,
	emailService EmailServiceClient,
	pushService PushServiceClient,
) *Server {
	return &Server{
		logger:       logger,
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
		grpc.UnaryInterceptor(s.loggingInterceptor),
	)
	RegisterNotificationServiceServer(s.grpcServer, s)

	s.logger.Info("Starting gRPC server", zap.String("port", port))
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
		case <-stopped:
		}
	}
}

func (s *Server) SendEmail(ctx context.Context, req *EmailRequest) (*NotificationResponse, error) {
	_, err := s.emailService.SendEmail(ctx, req)
	if err != nil {
		s.logger.Error("Failed to send email", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "email sending failed: %v", err)
	}

	return &NotificationResponse{
		Success:        true,
		Message:        "Email sent successfully",
		NotificationId: generateID(),
	}, nil
}

func (s *Server) SendPush(ctx context.Context, req *PushRequest) (*NotificationResponse, error) {
	_, err := s.pushService.SendPush(ctx, req)
	if err != nil {
		s.logger.Error("Failed to send push notification", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "push notification failed: %v", err)
	}

	return &NotificationResponse{
		Success:        true,
		Message:        "Push notification sent successfully",
		NotificationId: generateID(),
	}, nil
}

func (s *Server) loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp interface{}, err error) {
	start := time.Now()
	resp, err = handler(ctx, req)
	duration := time.Since(start)

	s.logger.Info("gRPC request",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", duration),
		zap.Error(err),
	)

	return resp, err
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
