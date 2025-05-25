package email

import (
	"context"
	"errors"
	"fmt"
	"net/smtp"
	"ride-sharing-notification/config"
	"ride-sharing-notification/internal/delivery/rpc"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	maxRetries      = 3
	retryDelay      = 1 * time.Second
	timeoutDuration = 10 * time.Second
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
		cfg.Email.SMTPHost,
	)

	return &Service{
		config: cfg,
		auth:   auth,
	}
}

func (s *Service) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *Service) SendEmail(ctx context.Context, req *rpc.EmailRequest) (*rpc.NotificationResponse, error) {
	if req == nil {
		return nil, errors.New("email request cannot be nil")
	}

	// Validate email fields
	if req.To == "" {
		return nil, errors.New("recipient email cannot be empty")
	}
	if req.Subject == "" || req.Body == "" {
		return nil, errors.New("email subject and body cannot be empty")
	}

	from := s.config.Email.FromEmail
	if from == "" {
		from = s.config.Email.Username
	}

	message := fmt.Sprintf(
		"From: %s\r\n"+
			"To: %s\r\n"+
			"Subject: %s\r\n"+
			"MIME-version: 1.0;\r\n"+
			"Content-Type: text/html; charset=\"UTF-8\";\r\n"+
			"\r\n"+
			"%s\r\n",
		from,
		req.To,
		req.Subject,
		req.Body,
	)

	var lastErr error

	// Retry logic
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
		defer cancel()

		err := s.sendEmail(ctx, from, req.To, []byte(message))
		if err == nil {
			// Success
			if s.logger != nil {
				s.logger.Info("Email sent successfully via Zoho Mail",
					zap.String("to", req.To),
					zap.String("subject", req.Subject),
					zap.Int("attempt", attempt),
				)
			}

			return &rpc.NotificationResponse{
				Success:        true,
				Message:        "Email sent successfully",
				NotificationId: generateID(),
			}, nil
		}

		lastErr = err
		if s.logger != nil {
			s.logger.Error("Failed to send email via Zoho Mail",
				zap.String("to", req.To),
				zap.String("subject", req.Subject),
				zap.Int("attempt", attempt),
				zap.Error(err),
			)
		}

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("after %d attempts, last error: %w", maxRetries, lastErr)
}

func (s *Service) sendEmail(ctx context.Context, from, to string, message []byte) error {
	done := make(chan error, 1)

	go func() {
		done <- smtp.SendMail(
			s.config.Email.SMTPHost+":"+s.config.Email.SMTPPort,
			s.auth,
			from,
			[]string{to},
			message,
		)
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func generateID() string {
	return uuid.New().String()
}
