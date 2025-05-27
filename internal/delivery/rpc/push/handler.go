package rpc

import (
	"context"
)

// EmailServiceClient is an interface for email service
type EmailServiceClient interface {
	SendEmail(ctx context.Context, req *EmailRequest) (*NotificationResponse, error)
}

// PushServiceClient is an interface for push notification service
type PushServiceClient interface {
	SendPush(ctx context.Context, req *PushRequest) (*NotificationResponse, error)
}
