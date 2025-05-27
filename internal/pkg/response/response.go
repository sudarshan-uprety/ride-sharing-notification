package response

import (
	"ride-sharing-notification/internal/proto/notification"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type ResponseBuilder struct {
	success bool
	message string
	code    codes.Code
}

func New() *ResponseBuilder {
	return &ResponseBuilder{}
}

func (rb *ResponseBuilder) Success() *ResponseBuilder {
	rb.success = true
	return rb
}

func (rb *ResponseBuilder) Error(code codes.Code) *ResponseBuilder {
	rb.success = false
	rb.code = code
	return rb
}

func (rb *ResponseBuilder) WithMessage(msg string) *ResponseBuilder {
	rb.message = msg
	return rb
}

func (rb *ResponseBuilder) WithData(payload proto.Message, meta *notification.MetaData) (*notification.StandardResponse, error) {
	data := &notification.DataResponse{}

	if payload != nil {
		anyPayload, err := anypb.New(payload)
		if err != nil {
			return nil, err
		}
		data.Payload = anyPayload
	}

	if meta != nil {
		data.Meta = meta
	}

	return &notification.StandardResponse{
		Success: true,
		Message: rb.message,
		Content: &notification.StandardResponse_Data{
			Data: data,
		},
	}, nil
}

func (rb *ResponseBuilder) WithError(errorCode string, details map[string]string) *notification.StandardResponse {
	return &notification.StandardResponse{
		Success: false,
		Message: rb.message,
		Content: &notification.StandardResponse_Error{
			Error: &notification.ErrorResponse{
				ErrorCode:    errorCode,
				ErrorMessage: rb.message,
				Details:      details,
			},
		},
	}
}

func (rb *ResponseBuilder) BuildStatusError() error {
	return status.Error(rb.code, rb.message)
}

func (rb *ResponseBuilder) SimpleSuccess() *notification.StandardResponse {
	return &notification.StandardResponse{
		Success: true,
		Message: rb.message,
	}
}

func (rb *ResponseBuilder) SimpleError(errorCode string) *notification.StandardResponse {
	return rb.WithError(errorCode, nil)
}
