package rpc

import (
	"context"
	"ride-sharing-notification/internal/proto/notification"
)

// EmailServiceClient is an interface for email service
type EmailServiceClient interface {
	SendRegisterEmail(ctx context.Context, req *notification.RegisterEmailRequest) (*notification.StandardResponse, error)
	SendForgetPasswordEmail(ctx context.Context, req *notification.ForgetPasswordEmailRequest) (*notification.StandardResponse, error)
}

type PushServiceClient interface {
	SendPush(ctx context.Context, req *notification.PushRequest) (*notification.StandardResponse, error)
}
