package grpc

import (
	"context"
	"net"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	pb.UnimplementedNotificationServiceServer
	kafkaProducer kafka.Producer
	logger        *zap.Logger
	emailService  pb.EmailServiceClient // Your email service interface
	pushService   pb.PushServiceClient  // Your push service interface
}

func NewServer(kafkaProducer kafka.Producer, logger *zap.Logger, emailService pb.EmailServiceClient, pushService pb.PushServiceClient) *Server {
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
		grpc.ConnectionTimeout(5 * time.Second),
	)
	pb.RegisterNotificationServiceServer(grpcServer, s)

	s.logger.Info("Starting gRPC server", zap.String("port", port))
	return grpcServer.Serve(lis)
}

func (s *Server) SendEmail(ctx context.Context, req *pb.EmailRequest) (*pb.NotificationResponse, error) {
	// Try direct send first
	resp, err := s.emailService.SendEmail(ctx, req)
	if err == nil {
		return &pb.NotificationResponse{
			Success:        true,
			Message:        "Email sent successfully",
			NotificationId: generateID(),
			UsedFallback:   false,
		}, nil
	}

	// If direct send fails and fallback is enabled
	if req.UseFallback {
		s.logger.Warn("Email sending failed, falling back to Kafka", zap.Error(err))

		kafkaReq := &pb.KafkaNotificationRequest{
			MessageId:        generateID(),
			Payload:          marshalToBytes(req),
			CreatedAt:        timestamppb.Now(),
			NotificationType: "email",
		}

		if err := s.kafkaProducer.Produce(ctx, "notification-fallback", kafkaReq); err != nil {
			return nil, status.Errorf(codes.Internal, "both direct send and fallback failed: %v", err)
		}

		return &pb.NotificationResponse{
			Success:        true,
			Message:        "Email queued in Kafka for later processing",
			NotificationId: kafkaReq.MessageId,
			UsedFallback:   true,
		}, nil
	}

	return nil, status.Errorf(codes.Internal, "email sending failed: %v", err)
}

// Similar implementation for SendPush would go here

func (s *Server) ProcessKafkaNotification(ctx context.Context, req *pb.KafkaNotificationRequest) (*pb.NotificationResponse, error) {
	switch req.NotificationType {
	case "email":
		var emailReq pb.EmailRequest
		if err := unmarshalFromBytes(req.Payload, &emailReq); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid email payload: %v", err)
		}
		return s.SendEmail(ctx, &emailReq)
	case "push":
		// Similar for push notifications
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown notification type: %s", req.NotificationType)
	}
}
