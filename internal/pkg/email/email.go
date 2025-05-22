package email

import (
	"context"
	"fmt"
	"net/smtp"
	"ride-sharing-notification/config"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service struct {
	config *config.Config
	logger *zap.Logger
	auth   smtp.Auth
}

func NewService(cfg *config.Config) *Service {
	auth := smtp.PlainAuth(
		"",
		cfg.Email.Username,
		cfg.Email.Password,
		"smtp.zoho.com",
	)

	return &Service{
		config: cfg,
		auth:   auth,
	}
}

func (s *Service) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *Service) SendEmail(ctx context.Context, req *EmailRequest) (*NotificationResponse, error) {
	from := s.config.Email.FromEmail
	if from == "" {
		from = s.config.Email.Username
	}

	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"\r\n"+
			"%s\r\n",
		from,
		req.To,
		req.Subject,
		req.Body,
	)

	err := smtp.SendMail(
		"smtp.zoho.com:587",
		s.auth,
		from,
		[]string{req.To},
		[]byte(message),
	)

	if err != nil {
		if s.logger != nil {
			s.logger.Error("Failed to send email via Zoho Mail",
				zap.String("to", req.To),
				zap.String("subject", req.Subject),
				zap.Error(err),
			)
		}
		return nil, err
	}

	if s.logger != nil {
		s.logger.Info("Email sent successfully via Zoho Mail",
			zap.String("to", req.To),
			zap.String("subject", req.Subject),
		)
	}

	return &NotificationResponse{
		Success:        true,
		Message:        "Email sent successfully",
		NotificationId: generateID(),
		UsedFallback:   false,
	}, nil
}

// generateID generates a unique ID for the notification
func generateID() string {
	// Implement your ID generation logic here
	// This is a simple placeholder implementation
	return uuid.New().String()
}
