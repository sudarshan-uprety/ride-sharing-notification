package kafka

import (
	"encoding/json"
	"log"
	"ride-sharing-notification/internal/pkg/email"

	"github.com/segmentio/kafka-go"
)

type Handler interface {
	Handle(msg kafka.Message) error
}

type MessageHandler struct {
	emailSvc *email.Service
}

func NewMessageHandler(emailSvc *email.Service) *MessageHandler {
	return &MessageHandler{
		emailSvc: emailSvc,
	}
}

func (h *MessageHandler) Handle(msg kafka.Message) error {
	var payload map[string]interface{}

	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return err
	}

	log.Printf("Sending OTP to: %s (otp: %s)", payload["to"], payload["otp"])

	emailPayload := &email.EmailPayload{
		To:         payload["to"].(string),
		EMAIL_TYPE: payload["type"].(string),
		Data:       payload,
	}
	log.Println(emailPayload)
	// h.emailSvc.VerifyEmail(context.Background(), emailPayload)

	return nil
}
