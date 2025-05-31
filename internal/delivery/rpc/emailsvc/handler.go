package emailsvc

import (
	"context"
	"ride-sharing-notification/internal/pkg/email"
	"ride-sharing-notification/internal/pkg/errors"
	"ride-sharing-notification/internal/pkg/response"
	"ride-sharing-notification/internal/proto/notification"
)

type Handler struct {
	emailService *email.Service
}

func NewHandler(emailService *email.Service) *Handler {
	return &Handler{emailService: emailService}
}

func (h *Handler) SendRegisterEmail(ctx context.Context, req *notification.RegisterEmailRequest) (*notification.StandardResponse, error) {
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

func (h *Handler) SendForgetPasswordEmail(ctx context.Context, req *notification.ForgetPasswordEmailRequest) (*notification.StandardResponse, error) {
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
