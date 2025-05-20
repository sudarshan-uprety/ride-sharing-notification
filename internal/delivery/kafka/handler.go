package kafka

import (
	"context"
	"time"

	"ride-sharing-notification/internal/delivery/rpc"

	"go.uber.org/zap"
)

type Handler struct {
	rpcServer *rpc.Server
	logger    *zap.Logger
}

func NewHandler(rpcServer *rpc.Server, logger *zap.Logger) *Handler {
	return &Handler{
		rpcServer: rpcServer,
		logger:    logger,
	}
}

func (h *Handler) Handle(ctx context.Context, message []byte) error {
	// Parse the Kafka message
	var kafkaReq pb.KafkaNotificationRequest
	if err := unmarshalFromBytes(message, &kafkaReq); err != nil {
		h.logger.Error("failed to unmarshal kafka message", zap.Error(err))
		return err
	}

	// Add retry logic with context timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Process the notification
	_, err := h.rpcServer.ProcessKafkaNotification(ctx, &kafkaReq)
	if err != nil {
		h.logger.Error("failed to process kafka notification",
			zap.String("notificationType", kafkaReq.NotificationType),
			zap.Error(err))
		return err
	}

	h.logger.Info("successfully processed kafka notification",
		zap.String("notificationType", kafkaReq.NotificationType),
		zap.String("messageId", kafkaReq.MessageId))
	return nil
}
