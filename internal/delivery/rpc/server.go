package rpc

import (
	"context"
	"net"
	"time"

	"ride-sharing-notification/internal/delivery/rpc/emailsvc"
	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/pkg/logging"
	"ride-sharing-notification/internal/pkg/middleware"
	"ride-sharing-notification/internal/proto/notification"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type GRPCServer struct {
	notification.UnimplementedNotificationServiceServer
	server        *grpc.Server
	healthServer  *health.Server
	emailHandler  *emailsvc.EmailServer
	shutdownGrace time.Duration
}

func NewGRPCServer(emailService *email.Service) *GRPCServer {
	return &GRPCServer{
		emailHandler:  emailsvc.NewEmailServer(emailService),
		shutdownGrace: 10 * time.Second,
	}
}

func (s *GRPCServer) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	s.server = grpc.NewServer(
		grpc.ConnectionTimeout(5*time.Second),
		grpc.ChainUnaryInterceptor(
			middleware.LoggingInterceptor(),
		),
	)

	s.healthServer = health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.server, s.healthServer)
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(s.server)

	notification.RegisterNotificationServiceServer(s.server, s.emailHandler)

	logging.GetLogger().Info("gRPC server starting",
		zap.String("port", port),
		zap.Duration("shutdown_grace_period", s.shutdownGrace),
	)

	return s.server.Serve(lis)
}

func (s *GRPCServer) Stop(ctx context.Context) {
	if s.server == nil {
		return
	}
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.shutdownGrace)
		defer cancel()
	}

	stopped := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-ctx.Done():
		logging.GetLogger().Warn("forcing gRPC server shutdown")
		s.server.Stop()
	case <-stopped:
		logging.GetLogger().Info("gRPC server stopped gracefully")
	}
}
