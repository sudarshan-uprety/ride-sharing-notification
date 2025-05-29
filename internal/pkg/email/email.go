package email

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/smtp"
	"path/filepath"
	"ride-sharing-notification/config"
	"ride-sharing-notification/internal/proto/notification"
	"time"

	"go.uber.org/zap"
)

const (
	maxRetries      = 3
	retryDelay      = 1 * time.Second
	timeoutDuration = 10 * time.Second
)

type Service struct {
	config      *config.Config
	logger      *zap.Logger
	auth        smtp.Auth
	templateDir string
}

func NewService(cfg *config.Config, templateDir string) *Service {
	auth := smtp.PlainAuth(
		"",
		cfg.Email.Username,
		cfg.Email.Password,
		cfg.Email.SMTPHost,
	)

	return &Service{
		config:      cfg,
		auth:        auth,
		templateDir: templateDir,
	}
}

func (s *Service) SetLogger(logger *zap.Logger) {
	s.logger = logger
}

func (s *Service) SendEmail(ctx context.Context, req *notification.EmailRequest) (*notification.StandardResponse, error) {
	if req == nil {
		return nil, errors.New("email request cannot be nil")
	}

	// Validate required fields
	if req.To == "" {
		return nil, errors.New("recipient email cannot be empty")
	}
	if req.EmailType == "" {
		return nil, errors.New("email type cannot be empty")
	}

	// Get template config
	templateConfig, exists := EmailTemplates[req.EmailType]
	if !exists {
		return nil, fmt.Errorf("unknown email type: %s", req.EmailType)
	}

	// Render template from file
	body, err := s.renderTemplate(templateConfig.TemplateFile, req.EmailType)
	if err != nil {
		return nil, fmt.Errorf("failed to render template: %w", err)
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
		templateConfig.Subject,
		body,
	)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(ctx, timeoutDuration)
		defer cancel()

		err := s.sendEmail(ctx, from, req.To, []byte(message))
		if err == nil {
			return &notification.StandardResponse{
				Success: true,
				Message: "Email sent successfully",
			}, nil
		}

		lastErr = err
		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	return nil, fmt.Errorf("after %d attempts, last error: %w", maxRetries, lastErr)
}

func (s *Service) renderTemplate(templateFile string, data string) (string, error) {
	templatePath := filepath.Join(s.templateDir, templateFile)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
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
