package errors

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "VALIDATION_ERROR"
	ErrorTypeVerification ErrorType = "VERIFICATION_ERROR"
	ErrorTypeConflict     ErrorType = "CONFLICT_ERROR"
	ErrorTypeNotFound     ErrorType = "NOT_FOUND_ERROR"
	ErrorTypeUnauthorized ErrorType = "UNAUTHORIZED_ERROR"
	ErrorTypeForbidden    ErrorType = "FORBIDDEN_ERROR"
	ErrorTypeInternal     ErrorType = "INTERNAL_ERROR"
)

type AppError struct {
	Type    ErrorType
	Message string
	Details any
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func toGRPCCode(t ErrorType) codes.Code {
	switch t {
	case ErrorTypeValidation, ErrorTypeVerification:
		return codes.InvalidArgument
	case ErrorTypeConflict:
		return codes.AlreadyExists
	case ErrorTypeNotFound:
		return codes.NotFound
	case ErrorTypeUnauthorized:
		return codes.Unauthenticated
	case ErrorTypeForbidden:
		return codes.PermissionDenied
	default:
		return codes.Internal
	}
}

// Converts AppError to gRPC status
func ToGRPCStatus(err *AppError) error {
	return status.Error(toGRPCCode(err.Type), err.Message)
}

// Factory functions
func NewValidationError(message string, details any) *AppError {
	return &AppError{
		Type:    ErrorTypeValidation,
		Message: message,
		Details: details,
	}
}

func NewVerificationError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeVerification,
		Message: message,
	}
}

func NewConflictError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeConflict,
		Message: message,
	}
}

func NewNotFoundError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeNotFound,
		Message: message,
	}
}

func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeUnauthorized,
		Message: message,
	}
}

func NewForbiddenError(message string) *AppError {
	return &AppError{
		Type:    ErrorTypeForbidden,
		Message: message,
	}
}

func NewInternalError(err error) *AppError {
	return &AppError{
		Type:    ErrorTypeInternal,
		Message: "internal server error",
		Err:     err,
	}
}
