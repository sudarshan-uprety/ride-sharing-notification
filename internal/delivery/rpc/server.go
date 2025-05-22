package rpc

import (
	"context"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

type Server struct {
	UnimplementedNotificationServiceServer
	kafkaProducer Producer
	logger        *zap.Logger
	emailService  EmailServiceClient
	pushService   PushServiceClient
}

// Producer defines the Kafka producer interface
type Producer interface {
	Produce(ctx context.Context, topic string, msg proto.Message) error
	Close() error
}

func NewServer(
	kafkaProducer Producer,
	logger *zap.Logger,
	emailService EmailServiceClient,
	pushService PushServiceClient,
) *Server {
	return &Server{
		kafkaProducer: kafkaProducer,
		logger:        logger,
		emailService:  emailService,
		pushService:   pushService,
	}
}

func (s *Server) Start(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.ConnectionTimeout(5*time.Second),
		grpc.UnaryInterceptor(s.loggingInterceptor),
	)
	RegisterNotificationServiceServer(grpcServer, s)

	s.logger.Info("Starting gRPC server", zap.String("port", port))
	return grpcServer.Serve(lis)
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
		zap.Any("request", req),
		zap.Error(err),
	)

	return resp, err
}
