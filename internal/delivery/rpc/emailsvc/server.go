package emailsvc

import (
	"context"

	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/proto/notification"

	"google.golang.org/grpc"
)

type EmailServer struct {
	notification.UnimplementedNotificationServiceServer
	handler *Handler
}

func NewEmailServer(emailService *email.Service) *EmailServer {
	return &EmailServer{
		handler: NewHandler(emailService),
	}
}

func Register(server *grpc.Server, emailService *email.Service) {
	notification.RegisterNotificationServiceServer(server, NewEmailServer(emailService))
}

func (s *EmailServer) SendRegisterEmail(ctx context.Context, req *notification.RegisterEmailRequest) (*notification.StandardResponse, error) {
	return s.handler.SendRegisterEmail(ctx, req)
}

func (s *EmailServer) SendForgetPasswordEmail(ctx context.Context, req *notification.ForgetPasswordEmailRequest) (*notification.StandardResponse, error) {
	return s.handler.SendForgetPasswordEmail(ctx, req)
}
