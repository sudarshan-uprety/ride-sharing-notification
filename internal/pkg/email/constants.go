package email

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"ride-sharing-notification/internal/constants"
	"ride-sharing-notification/internal/proto/notification"
)

type EmailService struct {
	templateDir string
}

func NewEmailService(templateDir string) *EmailService {
	return &EmailService{
		templateDir: templateDir,
	}
}

func (s *EmailService) SendEmail(req *notification.EmailRequest) error {
	// Get template config
	templateConfig, exists := constants.EmailTemplates[req.EmailType]
	if !exists {
		return fmt.Errorf("unknown email type: %s", req.EmailType)
	}

	// Read template file
	templatePath := filepath.Join(s.templateDir, templateConfig.TemplateFile)
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Execute template with dynamic data
	var body bytes.Buffer
	if err := tmpl.Execute(&body, req.TemplateData); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// Send email
	email := &Email{
		To:      req.To,
		Subject: templateConfig.Subject,
		Body:    body.String(),
	}

	return s.send(email)
}
