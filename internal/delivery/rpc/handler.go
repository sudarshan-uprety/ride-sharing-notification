package rpc

import (
	"context"
	"ride-sharing-notification/internal/proto/notification"
)

// EmailServiceClient is an interface for email service
type EmailServiceClient interface {
	SendEmail(ctx context.Context, req *notification.EmailRequest) (*notification.StandardResponse, error)
}

type PushServiceClient interface {
	SendPush(ctx context.Context, req *notification.PushRequest) (*notification.StandardResponse, error)
}
