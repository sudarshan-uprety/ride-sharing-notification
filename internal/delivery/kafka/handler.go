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

type MessageHandler struct{}

func NewMessageHandler() *MessageHandler {
	return &MessageHandler{}
}

func (h *MessageHandler) Handle(msg kafka.Message) error {
	var payload map[string]string

	if err := json.Unmarshal(msg.Value, &payload); err != nil {
		return err
	}

	switch payload["type"] {
	case email.EmailTypeRegister:
		log.Printf("Sending OTP to: %s (otp: %s)", payload["to"], payload["otp"])
	case email.EmailTypeForgetPassword:
		log.Printf("Sending OTP to: %s (otp: %s)", payload["to"], payload["otp"])
	default:
		log.Println("Unknown message type")
	}

	return nil
}
