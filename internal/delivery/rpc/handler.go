package rpc

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"google.golang.org/protobuf/proto"
)

func generateID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func marshalToBytes(msg proto.Message) []byte {
	b, _ := proto.Marshal(msg)
	return b
}

func unmarshalFromBytes(data []byte, msg proto.Message) error {
	return proto.Unmarshal(data, msg)
}

// EmailServiceClient is an interface for email service
type EmailServiceClient interface {
	SendEmail(ctx context.Context, req *EmailRequest) (*NotificationResponse, error)
}

// PushServiceClient is an interface for push notification service
type PushServiceClient interface {
	SendPush(ctx context.Context, req *PushRequest) (*NotificationResponse, error)
}
