package rpc

import (
	"context"
	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/pkg/errors"
	"ride-sharing-notification/internal/pkg/response"
	"ride-sharing-notification/internal/proto/notification"
)

// EmailHandler handles email-related gRPC requests
type EmailHandler struct {
	emailService *email.Service
}

func NewEmailHandler(emailService *email.Service) *EmailHandler {
	return &EmailHandler{
		emailService: emailService,
	}
}

func (h *EmailHandler) SendRegisterEmail(ctx context.Context, req *notification.RegisterEmailRequest) (*notification.StandardResponse, error) {
	// Validate request
	if err := validateEmailRequest(req); err != nil {
		return nil, errors.ToGRPCStatus(err)
	}

	// Build email payload
	payload := &email.EmailPayload{
		To:         req.To,
		EMAIL_TYPE: email.EmailTypeRegister,
		Data: map[string]interface{}{
			"name": req.To,
			"otp":  req.Otp,
		},
	}

	// Process the email
	emailResp, err := h.emailService.VerifyEmail(ctx, payload)
	if err != nil {
		appErr := errors.NewInternalError(err)
		return nil, errors.ToGRPCStatus(appErr)
	}

	// Build success response
	respBuilder := response.New().
		Success().
		WithMessage("Registration email sent successfully")

	if emailResp != nil {
		return respBuilder.WithData(emailResp, nil)
	}
	return respBuilder.SimpleSuccess(), nil
}

func (h *EmailHandler) SendForgetPasswordEmail(ctx context.Context, req *notification.ForgetPasswordEmailRequest) (*notification.StandardResponse, error) {
	// Validate request
	if err := validateEmailRequest(req); err != nil {
		return nil, errors.ToGRPCStatus(err)
	}

	// Build email payload
	payload := &email.EmailPayload{
		To:         req.To,
		EMAIL_TYPE: email.EmailTypeForgetPassword,
		Data: map[string]interface{}{
			"name": req.To,
			"otp":  req.Otp,
		},
	}

	// Process the email
	emailResp, err := h.emailService.VerifyEmail(ctx, payload)
	if err != nil {
		appErr := errors.NewInternalError(err)
		return nil, errors.ToGRPCStatus(appErr)
	}

	// Build success response
	respBuilder := response.New().
		Success().
		WithMessage("Password reset email sent successfully")

	if emailResp != nil {
		return respBuilder.WithData(emailResp, nil)
	}
	return respBuilder.SimpleSuccess(), nil
}

// validateEmailRequest validates common email request fields
func validateEmailRequest(req interface{}) *errors.AppError {
	switch r := req.(type) {
	case *notification.RegisterEmailRequest:
		if r.To == "" || r.Otp == "" {
			return errors.NewValidationError("invalid request", map[string]string{
				"to":  "required",
				"otp": "required",
			})
		}
	case *notification.ForgetPasswordEmailRequest:
		if r.To == "" || r.Otp == "" {
			return errors.NewValidationError("invalid request", map[string]string{
				"to":  "required",
				"otp": "required",
			})
		}
	default:
		return errors.NewValidationError("invalid request type", nil)
	}
	return nil
}
