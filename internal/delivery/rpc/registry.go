package rpc

import (
	"ride-sharing-notification/internal/delivery/rpc/emailsvc"
	"ride-sharing-notification/internal/pkg/email"
)

type Handlers struct {
	EmailHandler *emailsvc.Handler
	// PushHandler  *push.Handler
}

func RegisterAllHandlers(emailService *email.Service) *Handlers {
	return &Handlers{
		EmailHandler: emailsvc.NewHandler(emailService),
	}
}
