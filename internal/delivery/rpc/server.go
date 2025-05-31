package rpc

import (
	"context"
	"net"
	"ride-sharing-notification/internal/delivery/rpc/emailsvc"
	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/pkg/logging"
	"ride-sharing-notification/internal/pkg/middleware"
	"ride-sharing-notification/internal/proto/notification"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// GRPCServer manages the lifecycle of the gRPC server
type GRPCServer struct {
	notification.UnimplementedNotificationServiceServer
	server       *grpc.Server
	healthServer *health.Server
	emailHandler *emailsvc.Handler
	// pushHandler   *PushHandler
	shutdownGrace time.Duration
}

// NewGRPCServer creates a new gRPC server with all registered services
func NewGRPCServer(
	emailService *email.Service) *GRPCServer {
	s := &GRPCServer{
		shutdownGrace: 10 * time.Second, // Default grace period
	}

	// Initialize handlers
	s.emailHandler = emailsvc.NewHandler(emailService)
	// s.pushHandler = NewPushHandler(pushService)

	return s
}

// GRPCServerOption configures the GRPCServer
type GRPCServerOption func(*GRPCServer)

// WithShutdownGracePeriod sets the graceful shutdown period
func WithShutdownGracePeriod(d time.Duration) GRPCServerOption {
	return func(s *GRPCServer) {
		s.shutdownGrace = d
	}
}

// Start begins serving gRPC requests
func (s *GRPCServer) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// Configure server with interceptors
	s.server = grpc.NewServer(
		grpc.ConnectionTimeout(5*time.Second),
		grpc.ChainUnaryInterceptor(
			middleware.LoggingInterceptor(),
			// middleware.RecoveryInterceptor(),
			// middleware.AuthInterceptor(),
			// middleware.RequestIDInterceptor(),
		),
	)

	// Register services
	notification.RegisterNotificationServiceServer(s.server, s)

	// Register health service
	s.healthServer = health.NewServer()
	grpc_health_v1.RegisterHealthServer(s.server, s.healthServer)
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection for debugging
	reflection.Register(s.server)

	logging.GetLogger().Info("gRPC server starting",
		zap.String("port", port),
		zap.Duration("shutdown_grace_period", s.shutdownGrace),
	)

	return s.server.Serve(lis)
}

// Stop gracefully shuts down the server
func (s *GRPCServer) Stop(ctx context.Context) {
	if s.server == nil {
		return
	}

	// Set health status to NOT_SERVING
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	// Create a context with the grace period if none was provided
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

// SendRegisterEmail implements the NotificationServiceServer interface
func (s *GRPCServer) SendRegisterEmail(ctx context.Context, req *notification.RegisterEmailRequest) (*notification.StandardResponse, error) {
	return s.emailHandler.SendRegisterEmail(ctx, req)
}

// SendForgetPasswordEmail implements the NotificationServiceServer interface
func (s *GRPCServer) SendForgetPasswordEmail(ctx context.Context, req *notification.ForgetPasswordEmailRequest) (*notification.StandardResponse, error) {
	return s.emailHandler.SendForgetPasswordEmail(ctx, req)
}

// SendPush implements the NotificationServiceServer interface
// func (s *GRPCServer) SendPush(ctx context.Context, req *notification.PushRequest) (*notification.StandardResponse, error) {
// 	return s.pushHandler.SendPush(ctx, req)
// }
