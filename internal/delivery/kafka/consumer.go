package kafka

import (
	"context"
	"ride-sharing-notification/internal/delivery/rpc"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	handler *Handler
}

func NewConsumer(brokers []string, topic string, groupID string, logger *zap.Logger, rpcServer *rpc.Server) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
			MaxWait: 10 * time.Second,
		}),
		logger:  logger,
		handler: NewHandler(rpcServer, logger),
	}
}

func (c *Consumer) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return c.reader.Close()
		default:
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				c.logger.Error("failed to fetch message", zap.Error(err))
				continue
			}

			if err := c.handler.Handle(ctx, msg.Value); err != nil {
				c.logger.Error("failed to handle message",
					zap.ByteString("value", msg.Value),
					zap.Error(err))
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("failed to commit message",
					zap.ByteString("value", msg.Value),
					zap.Error(err))
			}
		}
	}
}
