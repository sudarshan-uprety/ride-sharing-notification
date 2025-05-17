package kafka

import (
	"context"
	"encoding/json"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type Consumer struct {
	reader      *kafka.Reader
	logger      *zap.Logger
	grpcClient  pb.NotificationServiceClient
	retryPolicy RetryPolicy
}

type RetryPolicy struct {
	MaxAttempts int
	Delay       time.Duration
}

func NewConsumer(brokers []string, topic string, groupID string, logger *zap.Logger, grpcClient pb.NotificationServiceClient) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
			GroupID: groupID,
		}),
		logger:     logger,
		grpcClient: grpcClient,
		retryPolicy: RetryPolicy{
			MaxAttempts: 3,
			Delay:       1 * time.Second,
		},
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

			var kafkaReq pb.KafkaNotificationRequest
			if err := json.Unmarshal(msg.Value, &kafkaReq); err != nil {
				c.logger.Error("failed to unmarshal kafka message", zap.Error(err))
				continue
			}

			// Process with retry
			for attempt := 1; attempt <= c.retryPolicy.MaxAttempts; attempt++ {
				_, err := c.grpcClient.ProcessKafkaNotification(ctx, &kafkaReq)
				if err == nil {
					break
				}

				if attempt == c.retryPolicy.MaxAttempts {
					c.logger.Error("max retries exceeded for kafka message",
						zap.String("message_id", kafkaReq.MessageId),
						zap.Error(err))
					// TODO: Move to dead letter queue
				}

				time.Sleep(c.retryPolicy.Delay * time.Duration(attempt))
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("failed to commit message", zap.Error(err))
			}
		}
	}
}
