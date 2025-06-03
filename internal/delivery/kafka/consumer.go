package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	reader  *kafka.Reader
	handler *Handler
}

func NewConsumer(brokers []string, topic string, groupID string) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			GroupID: groupID,
			Topic:   topic,
			MaxWait: 10 * time.Second,
		}),
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
				continue
			}

			if err := c.handler.Handle(ctx, msg.Value); err != nil {
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
			}
		}
	}
}
