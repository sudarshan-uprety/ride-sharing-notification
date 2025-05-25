package response

import (
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

func (rb *ResponseBuilder) WithData(payload proto.Message, meta *MetaData) (*StandardResponse, error) {
	data := &DataResponse{}

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

	return &StandardResponse{
		Success: rb.success,
		Message: rb.message,
		Content: &StandardResponse_Data{
			Data: data,
		},
	}, nil
}

func (rb *ResponseBuilder) WithErrorDetails(errorCode string, details map[string]string) (*StandardResponse, error) {
	return &StandardResponse{
		Success: false,
		Message: rb.message,
		Content: &StandardResponse_Error{
			Error: &ErrorResponse{
				ErrorCode:    errorCode,
				ErrorMessage: rb.message,
				Details:      details,
			},
		},
	}, nil
}

func (rb *ResponseBuilder) BuildStatusError() error {
	return status.Error(rb.code, rb.message)
}
