package emailsvc

import (
	"ride-sharing-notification/internal/pkg/errors"
	"ride-sharing-notification/internal/proto/notification"
)

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
