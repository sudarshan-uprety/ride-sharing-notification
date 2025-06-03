package kafka

import (
	"encoding/json"
	"log"

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

	log.Printf("[Kafka] Message received: %v\n", payload)

	// Example: route by "type"
	switch payload["type"] {
	case "otp-forget-password":
		log.Printf("Sending OTP to: %s (otp: %s)", payload["to"], payload["otp"])
	default:
		log.Println("Unknown message type")
	}

	return nil
}
