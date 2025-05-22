package email

import (
	"context"
	"fmt"
	"net/smtp"
	"ride-sharing-notification/configs"

	"go.uber.org/zap"
)

// Service implements the EmailServiceClient interface
type Service struct {
	config *configs.EmailConfig
	logger *zap.Logger
	auth   smtp.Auth
}

// NewService creates a new email service with Zoho Mail configuration
func NewService(cfg *configs.EmailConfig) *Service {
	// Set up SMTP authentication for Zoho Mail
	auth := smtp.PlainAuth(
		"",
		cfg.Username, // Your Zoho Mail email address
		cfg.Password, // Your Zoho Mail password or app-specific password
		"smtp.zoho.com",
	)

	return &Service{
		config: cfg,
		auth:   auth,
	}
}

// SetLogger sets the logger for the email service
func (s *Service) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

// SendEmail sends an email using Zoho Mail SMTP
func (s *Service) SendEmail(ctx context.Context, req *EmailRequest) (*NotificationResponse, error) {
	// Construct the email message
	from := s.config.FromEmail
	if from == "" {
		from = s.config.Username // Fallback to username if from email not specified
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

	// Send the email using Zoho's SMTP server
	err := smtp.SendMail(
		"smtp.zoho.com:587", // Zoho Mail SMTP server with TLS
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
		Message:        "Email sent